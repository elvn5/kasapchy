// Package httpapi exposes the users use cases over HTTP (Gin).
package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/your-org/app/pkg/httpserver"
	"github.com/your-org/app/users/internal/domain"
)

// UserUseCase is the application logic the handler depends on.
type UserUseCase interface {
	List(ctx context.Context) ([]domain.User, error)
	Get(ctx context.Context, id string) (domain.User, error)
	RegisterFromTelegram(ctx context.Context, profile domain.TelegramProfile) (domain.User, error)
}

// Handler serves the users routes.
//
//   - /users/*          — admin API, reached through the gateway (which checks the session);
//   - /internal/users/* — service-to-service API (bot), never exposed by the gateway.
type Handler struct {
	users UserUseCase
	log   *slog.Logger
}

// NewHandler creates a Handler.
func NewHandler(users UserUseCase, log *slog.Logger) *Handler {
	return &Handler{users: users, log: log}
}

// Register mounts the routes on r.
func (h *Handler) Register(r gin.IRouter) {
	r.GET("/users", h.list)
	r.GET("/users/:id", h.get)
	r.POST("/internal/users/telegram", h.registerFromTelegram)
}

type userResponse struct {
	ID         string    `json:"id"`
	TelegramID int64     `json:"telegramId,string"`
	Username   *string   `json:"username"`
	FirstName  *string   `json:"firstName"`
	LastName   *string   `json:"lastName"`
	Phone      *string   `json:"phone"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

func toResponse(u domain.User) userResponse {
	return userResponse{
		ID:         u.ID,
		TelegramID: u.TelegramID,
		Username:   u.Username,
		FirstName:  u.FirstName,
		LastName:   u.LastName,
		Phone:      u.Phone,
		CreatedAt:  u.CreatedAt,
		UpdatedAt:  u.UpdatedAt,
	}
}

func (h *Handler) list(c *gin.Context) {
	users, err := h.users.List(c.Request.Context())
	if err != nil {
		h.internalError(c, err)
		return
	}

	resp := make([]userResponse, 0, len(users))
	for _, u := range users {
		resp = append(resp, toResponse(u))
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) get(c *gin.Context) {
	user, err := h.users.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			httpserver.Error(c, http.StatusNotFound, "User not found")
			return
		}
		h.internalError(c, err)
		return
	}
	c.JSON(http.StatusOK, toResponse(user))
}

type registerFromTelegramRequest struct {
	TelegramID int64   `json:"telegramId" binding:"required,gt=0"`
	Username   *string `json:"username" binding:"omitempty,max=255"`
	FirstName  *string `json:"firstName" binding:"omitempty,max=255"`
	LastName   *string `json:"lastName" binding:"omitempty,max=255"`
}

func (h *Handler) registerFromTelegram(c *gin.Context) {
	var req registerFromTelegramRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.users.RegisterFromTelegram(c.Request.Context(), domain.TelegramProfile{
		TelegramID: req.TelegramID,
		Username:   req.Username,
		FirstName:  req.FirstName,
		LastName:   req.LastName,
	})
	if err != nil {
		h.internalError(c, err)
		return
	}
	c.JSON(http.StatusOK, toResponse(user))
}

func (h *Handler) internalError(c *gin.Context, err error) {
	h.log.ErrorContext(c.Request.Context(), "request failed", slog.Any("error", err))
	httpserver.Error(c, http.StatusInternalServerError, "Internal server error")
}
