package handlers

import (
	"context"
	"net/http"
	"time"

	"wunderlist-backend/middleware"
	"wunderlist-backend/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateList
// @Summary Create a new list
// @Description Create a new task list
// @Tags Lists
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param list body models.List true "List info"
// @Success 201 {object} models.MessageResponse
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

	userID := c.GetString(middleware.UserIDKey)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Unauthorized"})
		return
	}

	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid user ID"})
		return
	}

	list.ID = primitive.NewObjectID()
	list.UserID = uid
	list.CreatedAt = time.Now().UTC()
	list.UpdatedAt = list.CreatedAt

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	if _, err := ListCollection.InsertOne(ctx, list); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to create list"})
		return
	}

	c.JSON(http.StatusCreated, models.MessageResponse{Message: "List created"})
}

// GetLists
// @Summary Get all lists
// @Description Retrieve all lists for the logged-in user
// @Tags Lists
// @Security BearerAuth
// @Produce json
// @Success 200 {array} models.List
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /lists [get]
func GetLists(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Unauthorized"})
		return
	}

	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid user ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	cursor, err := ListCollection.Find(ctx, bson.M{"user_id": uid})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch lists"})
		return
	}
	defer cursor.Close(ctx)

	var lists []models.List
	if err := cursor.All(ctx, &lists); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to parse lists"})
		return
	}

	c.JSON(http.StatusOK, lists)
}

// UpdateList
// @Summary Update a list
// @Description Update list title
// @Tags Lists
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "List ID"
// @Param list body models.List true "Updated list info"
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
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid list ID"})
		return
	}

	userID := c.GetString(middleware.UserIDKey)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Unauthorized"})
		return
	}

	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid user ID"})
		return
	}

	var body struct {
		Title string `json:"title"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
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
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to update list"})
		return
	}
	if res.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "List not found"})
		return
	}

	var updated models.List
	if err := ListCollection.FindOne(ctx, bson.M{"_id": listOID, "user_id": uid}).Decode(&updated); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch updated list"})
		return
	}

	c.JSON(http.StatusOK, updated)
}

// DeleteList
// @Summary Delete a list
// @Description Delete a list by ID
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
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid list ID"})
		return
	}

	userID := c.GetString(middleware.UserIDKey)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Unauthorized"})
		return
	}

	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid user ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	res, err := ListCollection.DeleteOne(ctx, bson.M{"_id": listOID, "user_id": uid})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to delete list"})
		return
	}
	if res.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "List not found"})
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{Message: "List deleted"})
}
