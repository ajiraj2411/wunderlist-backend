package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/ajiraj2411/wunderlist-backend/internal/auth"
	"github.com/ajiraj2411/wunderlist-backend/internal/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ListSessions
// @Summary List active sessions (cursor paginated)
// @Tags Sessions
// @Security BearerAuth
// @Produce json
// @Param cursor query string false "Pagination cursor"
// @Param limit query int false "Page size (default 20, max 100)"
// @Success 200 {object} models.CursorPage[models.SessionResponse]
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /sessions [get]
func ListSessions(c *gin.Context) {

	userID := c.GetString(UserIDKey)
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "invalid user"})
		return
	}

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

	cur, err := SessionCollection.Find(ctx, filter, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "db error"})
		return
	}
	defer cur.Close(ctx)

	var sessions []models.Session
	if err := cur.All(ctx, &sessions); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "parse error"})
		return
	}

	hasMore := false
	if len(sessions) > limit {
		hasMore = true
		sessions = sessions[:limit]
	}

	currentSessionHex := c.GetString("sessionID")
	currentSessionID, _ := primitive.ObjectIDFromHex(currentSessionHex)

	resp := make([]models.SessionResponse, 0, len(sessions))
	for _, s := range sessions {
		resp = append(resp, models.NewSessionResponse(
			s,
			currentSessionID,
		))
	}

	nextCursor := ""
	if hasMore && len(sessions) > 0 {
		last := sessions[len(sessions)-1]
		nextCursor, _ = models.EncodeCursor(last.CreatedAt, last.ID, uid)
	}

	c.JSON(http.StatusOK, models.CursorPage[models.SessionResponse]{
		Items:      resp,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	})
}

// @Summary Revoke a session
// @Tags Sessions
// @Security BearerAuth
// @Param id path string true "Session ID"
// @Success 204
// @Router /api/sessions/{id} [delete]
func RevokeSession(c *gin.Context) {

	userIDHex := c.GetString("userID")
	userID, _ := primitive.ObjectIDFromHex(userIDHex)

	sessionID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "invalid session id",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := auth.DeleteSessionByID(ctx, sessionID, userID)
	if err != nil || res == 0 {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "session not found",
		})
		return
	}

	c.Status(http.StatusNoContent)
}
