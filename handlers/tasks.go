package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"wunderlist-backend/internal/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//
// CREATE TASK
//

// CreateTask
// @Summary Create a task
// @Description Add a task to a list (user-owned)
// @Tags Tasks
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param task body models.Task true "Task info"
// @Success 201 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /tasks [post]
func CreateTask(c *gin.Context) {

	var task models.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	userID := c.GetString(UserIDKey)
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "invalid user id"})
		return
	}

	if task.Title == "" || task.ListID.IsZero() {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "title and list_id are required",
		})
		return
	}

	now := time.Now().UTC()
	task.ID = primitive.NewObjectID()
	task.UserID = uid
	task.Completed = false
	task.CreatedAt = now
	task.UpdatedAt = now

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	// Prevent duplicate titles per list per user
	count, err := TaskCollection.CountDocuments(ctx, bson.M{
		"user_id": uid,
		"list_id": task.ListID,
		"title":   task.Title,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "failed to validate task",
		})
		return
	}
	if count > 0 {
		c.JSON(http.StatusConflict, models.ErrorResponse{
			Error: "task with same title already exists in this list",
		})
		return
	}

	if _, err := TaskCollection.InsertOne(ctx, task); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "failed to create task",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":      task.ID.Hex(),
		"message": "task created",
	})
}

//
// GET TASKS
//

// GetTasks
// @Summary Get tasks
// @Description Retrieve tasks with optional filters and pagination
// @Tags Tasks
// @Security BearerAuth
// @Produce json
// @Param list_id query string false "List ID"
// @Param done query bool false "Completion status"
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Success 200 {array} models.Task
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /tasks [get]
func GetTasks(c *gin.Context) {

	userID := c.GetString(UserIDKey)
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "invalid user id"})
		return
	}

	filter := bson.M{"user_id": uid}

	if listID := c.Query("list_id"); listID != "" {
		lid, err := primitive.ObjectIDFromHex(listID)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid list_id"})
			return
		}
		filter["list_id"] = lid
	}

	if doneStr := c.Query("done"); doneStr != "" {
		done, err := strconv.ParseBool(doneStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid done value"})
			return
		}
		filter["completed"] = done
	}

	page := maxInt(parseInt(c.DefaultQuery("page", "1")), 1)
	limit := minInt(maxInt(parseInt(c.DefaultQuery("limit", "10")), 1), 100)
	skip := (page - 1) * limit

	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	cursor, err := TaskCollection.Find(ctx, filter, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "db error"})
		return
	}

	var tasks []models.Task
	if err := cursor.All(ctx, &tasks); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "parse error"})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

//
// ACTIVE TASKS
//

// GetActiveTasks
// @Summary Get active tasks
// @Description Retrieve incomplete tasks for the logged-in user
// @Tags Tasks
// @Security BearerAuth
// @Produce json
// @Param list_id query string false "List ID"
// @Success 200 {array} models.Task
// @Router /tasks/active [get]
func GetActiveTasks(c *gin.Context) {

	userID := c.GetString(UserIDKey)
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "invalid user id"})
		return
	}

	filter := bson.M{
		"user_id":   uid,
		"completed": false,
	}

	if listID := c.Query("list_id"); listID != "" {
		lid, err := primitive.ObjectIDFromHex(listID)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid list_id"})
			return
		}
		filter["list_id"] = lid
	}

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	cursor, err := TaskCollection.Find(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "db error"})
		return
	}

	var tasks []models.Task
	cursor.All(ctx, &tasks)

	c.JSON(http.StatusOK, tasks)
}

//
// SEARCH TASKS
//

// SearchTasks
// @Summary Search tasks
// @Description Text search tasks by keyword
// @Tags Tasks
// @Security BearerAuth
// @Produce json
// @Param q query string true "Search query"
// @Success 200 {array} models.Task
// @Router /tasks/search [get]
func SearchTasks(c *gin.Context) {

	userID := c.GetString(UserIDKey)
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "missing q"})
		return
	}

	uid, _ := primitive.ObjectIDFromHex(userID)

	filter := bson.M{
		"user_id": uid,
		"$text":   bson.M{"$search": query},
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "score", Value: bson.M{"$meta": "textScore"}}})

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	cursor, err := TaskCollection.Find(ctx, filter, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "db error"})
		return
	}

	var tasks []models.Task
	cursor.All(ctx, &tasks)

	c.JSON(http.StatusOK, tasks)
}

//
// UPDATE TASK
//

// UpdateTask
// @Summary Update a task
// @Description Update allowed task fields
// @Tags Tasks
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Param task body map[string]interface{} true "Fields to update"
// @Success 200 {object} models.Task
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /tasks/{id} [put]
func UpdateTask(c *gin.Context) {

	taskID := c.Param("id")
	id, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid task id"})
		return
	}

	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	allowed := bson.M{}
	for _, k := range []string{"title", "description", "completed", "priority", "due_date", "list_id"} {
		if v, ok := body[k]; ok {
			allowed[k] = v
		}
	}

	if len(allowed) == 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "no valid fields"})
		return
	}

	allowed["updated_at"] = time.Now().UTC()

	userID := c.GetString(UserIDKey)
	uid, _ := primitive.ObjectIDFromHex(userID)

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	res, err := TaskCollection.UpdateOne(ctx,
		bson.M{"_id": id, "user_id": uid},
		bson.M{"$set": allowed},
	)
	if err != nil || res.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "task not found"})
		return
	}

	var updated models.Task
	TaskCollection.FindOne(ctx,
		bson.M{"_id": id, "user_id": uid},
	).Decode(&updated)

	c.JSON(http.StatusOK, updated)
}

//
// DELETE TASK
//

// DeleteTask
// @Summary Delete a task
// @Description Delete a task by ID
// @Tags Tasks
// @Security BearerAuth
// @Produce json
// @Param id path string true "Task ID"
// @Success 200 {object} models.MessageResponse
// @Router /tasks/{id} [delete]
func DeleteTask(c *gin.Context) {

	taskID := c.Param("id")
	id, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid task id"})
		return
	}

	userID := c.GetString(UserIDKey)
	uid, _ := primitive.ObjectIDFromHex(userID)

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	res, err := TaskCollection.DeleteOne(ctx,
		bson.M{"_id": id, "user_id": uid})
	if err != nil || res.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "task not found"})
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{Message: "task deleted"})
}

//
// helpers
//

func parseInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
