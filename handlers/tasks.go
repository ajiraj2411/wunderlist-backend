package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"wunderlist-backend/middleware"
	"wunderlist-backend/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CreateTask
// @Summary Create a task
// @Description Add a task to a list (user-specific)
// @Tags Tasks
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param task body models.Task true "Task info"
// @Success 201 {object} models.MessageResponse "Task created"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 409 {object} models.ErrorResponse "Conflict: task exists"
// @Failure 500 {object} models.ErrorResponse "Failed to create task"
// @Router /tasks [post]
func CreateTask(c *gin.Context) {
	var task models.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	userID := c.GetString(middleware.UserIDKey)
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid user ID"})
		return
	}
	task.UserID = uid

	if task.ListID.IsZero() {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "list_id is required"})
		return
	}

	now := time.Now().UTC()
	task.CreatedAt = now
	task.UpdatedAt = now

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	// Prevent duplicate task titles in the same list for this user
	count, err := TaskCollection.CountDocuments(ctx, bson.M{
		"user_id": uid,
		"list_id": task.ListID,
		"title":   task.Title,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to check existing task"})
		return
	}
	if count > 0 {
		c.JSON(http.StatusConflict, models.ErrorResponse{Error: "Task with same title already exists in this list"})
		return
	}

	if _, err := TaskCollection.InsertOne(ctx, task); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to create task"})
		return
	}

	c.JSON(http.StatusCreated, models.MessageResponse{Message: "Task created"})
}

// GetTasks
// @Summary Get tasks
// @Description Retrieve all tasks, optionally filtered by list or completion status, with pagination
// @Tags Tasks
// @Security BearerAuth
// @Produce json
// @Param list_id query string false "List ID"
// @Param done query bool false "Filter by completion status"
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Page size (default 10)"
// @Success 200 {array} models.Task
// @Failure 400 {object} models.ErrorResponse "Invalid parameters"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /tasks [get]
func GetTasks(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid user ID"})
		return
	}

	filter := bson.M{"user_id": uid}

	// Optional list filter
	if listID := c.Query("list_id"); listID != "" {
		lid, err := primitive.ObjectIDFromHex(listID)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid list_id"})
			return
		}
		filter["list_id"] = lid
	}

	// Optional done filter
	if doneStr := c.Query("done"); doneStr != "" {
		done, err := strconv.ParseBool(doneStr)
		if err == nil {
			filter["completed"] = done
		}
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit < 1 {
		limit = 10
	}
	skip := (page - 1) * limit

	opts := options.Find().SetSkip(int64(skip)).SetLimit(int64(limit))

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	cursor, err := TaskCollection.Find(ctx, filter, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "DB error"})
		return
	}

	var tasks []models.Task
	if err := cursor.All(ctx, &tasks); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to parse tasks"})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

// GetActiveTasks
// @Summary Get active tasks
// @Description Retrieve all tasks that are not completed for the logged-in user, with optional list filter and pagination
// @Tags Tasks
// @Security BearerAuth
// @Produce json
// @Param list_id query string false "Optional List ID"
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Page size (default 10)"
// @Success 200 {array} models.Task
// @Failure 400 {object} models.ErrorResponse "Invalid parameters"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /tasks/active [get]
func GetActiveTasks(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid user ID"})
		return
	}

	filter := bson.M{"user_id": uid, "completed": false}

	if listID := c.Query("list_id"); listID != "" {
		lid, err := primitive.ObjectIDFromHex(listID)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid list_id"})
			return
		}
		filter["list_id"] = lid
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit < 1 {
		limit = 10
	}
	skip := (page - 1) * limit

	opts := options.Find().SetSkip(int64(skip)).SetLimit(int64(limit))

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	cursor, err := TaskCollection.Find(ctx, filter, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "DB error"})
		return
	}

	var tasks []models.Task
	if err := cursor.All(ctx, &tasks); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to parse tasks"})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

// SearchTasks
// @Summary Search tasks
// @Description Search tasks by title/description keyword, optional list filter and pagination
// @Tags Tasks
// @Security BearerAuth
// @Produce json
// @Param q query string true "Search query"
// @Param list_id query string false "Optional List ID"
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Success 200 {array} models.Task
// @Failure 400 {object} models.ErrorResponse "Missing query or invalid parameters"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /tasks/search [get]
func SearchTasks(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	query := c.Query("q")
	listID := c.Query("list_id")
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")

	if query == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Missing query"})
		return
	}

	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid user ID"})
		return
	}

	filter := bson.M{"user_id": uid, "$text": bson.M{"$search": query}}
	if listID != "" {
		lid, err := primitive.ObjectIDFromHex(listID)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid list_id"})
			return
		}
		filter["list_id"] = lid
	}

	p, _ := strconv.Atoi(page)
	if p < 1 {
		p = 1
	}
	l, _ := strconv.Atoi(limit)
	if l < 1 {
		l = 10
	}
	skip := (p - 1) * l

	opts := options.Find().
		SetSort(bson.D{{Key: "score", Value: bson.M{"$meta": "textScore"}}}).
		SetSkip(int64(skip)).
		SetLimit(int64(l))

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	cursor, err := TaskCollection.Find(ctx, filter, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "DB error"})
		return
	}

	var tasks []models.Task
	if err := cursor.All(ctx, &tasks); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to parse DB results"})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

// UpdateTask
// @Summary Update a task
// @Description Update task title, list, or completion status (user ownership enforced)
// @Tags Tasks
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Param task body models.Task true "Updated task info"
// @Success 200 {object} models.Task
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 404 {object} models.ErrorResponse "Task not found or not owned by user"
// @Failure 500 {object} models.ErrorResponse "Failed to update task"
// @Router /tasks/{id} [put]
func UpdateTask(c *gin.Context) {
	taskID := c.Param("id")
	id, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid task ID"})
		return
	}

	var update bson.M
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	// whitelist only allowed fields
	allowed := bson.M{}
	for _, k := range []string{"title", "description", "completed", "list_id"} {
		if v, ok := update[k]; ok {
			allowed[k] = v
		}
	}
	if len(allowed) == 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "No valid fields to update"})
		return
	}
	allowed["updated_at"] = time.Now().UTC()

	userID := c.GetString(middleware.UserIDKey)
	uid, _ := primitive.ObjectIDFromHex(userID)

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	res, err := TaskCollection.UpdateOne(ctx, bson.M{"_id": id, "user_id": uid}, bson.M{"$set": allowed})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to update task"})
		return
	}
	if res.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Task not found or not owned by user"})
		return
	}

	var task models.Task
	if err := TaskCollection.FindOne(ctx, bson.M{"_id": id, "user_id": uid}).Decode(&task); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch updated task"})
		return
	}

	c.JSON(http.StatusOK, task)
}

// DeleteTask
// @Summary Delete a task
// @Description Delete a task by ID (user ownership enforced)
// @Tags Tasks
// @Security BearerAuth
// @Produce json
// @Param id path string true "Task ID"
// @Success 200 {object} models.MessageResponse "Task deleted"
// @Failure 400 {object} models.ErrorResponse "Invalid task ID"
// @Failure 404 {object} models.ErrorResponse "Task not found or not owned by user"
// @Failure 500 {object} models.ErrorResponse "Failed to delete task"
// @Router /tasks/{id} [delete]
func DeleteTask(c *gin.Context) {
	taskID := c.Param("id")
	id, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid task ID"})
		return
	}

	userID := c.GetString(middleware.UserIDKey)
	uid, _ := primitive.ObjectIDFromHex(userID)

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	res, err := TaskCollection.DeleteOne(ctx, bson.M{"_id": id, "user_id": uid})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to delete task"})
		return
	}
	if res.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Task not found or not owned by user"})
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{Message: "Task deleted"})
}
