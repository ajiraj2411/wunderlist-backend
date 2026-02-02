package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/ajiraj2411/wunderlist-backend/internal/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CreateList
// @Summary Create a new list
// @Description Create a new task list for the logged-in user
// @Tags Lists
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param list body models.List true "List info"
// @Success 201 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /lists [post]
func CreateList(c *gin.Context) {

	var list models.List
	if err := c.ShouldBindJSON(&list); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	userID := c.GetString(UserIDKey)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "unauthorized"})
		return
	}

	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "invalid user id"})
		return
	}

	now := time.Now().UTC()
	list.ID = primitive.NewObjectID()
	list.UserID = uid
	list.CreatedAt = now
	list.UpdatedAt = now

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	if _, err := ListCollection.InsertOne(ctx, list); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "failed to create list",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":      list.ID.Hex(),
		"message": "list created",
	})
}

// GetLists
// @Summary Get lists (cursor paginated)
// @Description Retrieve lists for the logged-in user with cursor pagination
// @Tags Lists
// @Security BearerAuth
// @Produce json
// @Param cursor query string false "Pagination cursor"
// @Param limit query int false "Page size (default 20, max 100)"
// @Success 200 {object} models.CursorPage[models.List]
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /lists [get]
func GetLists(c *gin.Context) {

	userID := c.GetString(UserIDKey)
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "invalid user id"})
		return
	}

	// limit
	limit := maxInt(parseInt(c.DefaultQuery("limit", "20")), 1)
	limit = minInt(limit, 100)

	filter := bson.M{"user_id": uid}

	// cursor
	if cur := c.Query("cursor"); cur != "" {
		curCreatedAt, curID, err := models.DecodeCursor(cur, uid)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid cursor"})
			return
		}

		filter["$or"] = []bson.M{
			{"created_at": bson.M{"$lt": curCreatedAt}},
			{"created_at": curCreatedAt, "_id": bson.M{"$lt": curID}},
		}
	}

	opts := options.Find().
		SetSort(bson.D{
			{Key: "created_at", Value: -1},
			{Key: "_id", Value: -1},
		}).
		SetLimit(int64(limit + 1))

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	cursor, err := ListCollection.Find(ctx, filter, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "failed to fetch lists",
		})
		return
	}
	defer cursor.Close(ctx)

	var lists []models.List
	if err := cursor.All(ctx, &lists); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "failed to parse lists",
		})
		return
	}

	hasMore := false
	if len(lists) > limit {
		hasMore = true
		lists = lists[:limit]
	}

	nextCursor := ""
	if hasMore && len(lists) > 0 {
		last := lists[len(lists)-1]
		nextCursor, _ = models.EncodeCursor(last.CreatedAt, last.ID, uid)
	}

	c.JSON(http.StatusOK, models.CursorPage[models.List]{
		Items:      lists,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	})
}

// UpdateList
// @Summary Update a list
// @Description Update list title (user-owned)
// @Tags Lists
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "List ID"
// @Param list body map[string]string true "Updated list info"
// @Success 200 {object} models.List
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /lists/{id} [put]
func UpdateList(c *gin.Context) {

	listID := c.Param("id")
	listOID, err := primitive.ObjectIDFromHex(listID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid list id"})
		return
	}

	userID := c.GetString(UserIDKey)
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "invalid user id"})
		return
	}

	var body struct {
		Title string `json:"title"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Title == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "title is required"})
		return
	}

	update := bson.M{
		"title":      body.Title,
		"updated_at": time.Now().UTC(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	res, err := ListCollection.UpdateOne(
		ctx,
		bson.M{"_id": listOID, "user_id": uid},
		bson.M{"$set": update},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "update failed"})
		return
	}
	if res.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "list not found"})
		return
	}

	var updated models.List
	if err := ListCollection.FindOne(ctx,
		bson.M{"_id": listOID, "user_id": uid},
	).Decode(&updated); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "fetch failed"})
		return
	}

	c.JSON(http.StatusOK, updated)
}

// DeleteList
// @Summary Delete a list
// @Description Delete a list by ID (user-owned)
// @Tags Lists
// @Security BearerAuth
// @Produce json
// @Param id path string true "List ID"
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /lists/{id} [delete]
func DeleteList(c *gin.Context) {

	listID := c.Param("id")
	listOID, err := primitive.ObjectIDFromHex(listID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid list id"})
		return
	}

	userID := c.GetString(UserIDKey)
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "invalid user id"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	res, err := ListCollection.DeleteOne(ctx,
		bson.M{"_id": listOID, "user_id": uid})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "delete failed"})
		return
	}
	if res.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "list not found"})
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{
		Message: "list deleted",
	})
}
