package handlers

import (
	"context"
	"net/http"
	"wunderlist-backend/internal/auth"
	"wunderlist-backend/internal/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// @Summary Admin Force Logout User
// @Tags Admin Auth
// @Accept json
// @Produce json
// @Param userID path string true "User ID"
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/users/{userID}/force-logout [post]
func AdminForceLogoutUser(c *gin.Context) {
	userID, err := primitive.ObjectIDFromHex(c.Param("userID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid user ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	if err := auth.DeleteAllSessions(ctx, userID); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to force logout"})
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{
		Message: "user logged out from all devices",
	})
}
