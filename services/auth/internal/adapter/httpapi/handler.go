// Package httpapi exposes the auth use cases over HTTP (Gin).
package httpapi

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/your-org/app/auth/internal/domain"
	"github.com/your-org/app/pkg/authtoken"
	"github.com/your-org/app/pkg/httpserver"
)

// AuthUseCase is the application logic the handler depends on.
type AuthUseCase interface {
	Login(password string) (domain.Token, error)
	Authenticate(token string) (domain.Session, error)
}

// Handler serves /auth/* routes.
type Handler struct {
	auth         AuthUseCase
	log          *slog.Logger
	secureCookie bool
}

// NewHandler creates a Handler. secureCookie should be true behind HTTPS (production).
func NewHandler(auth AuthUseCase, log *slog.Logger, secureCookie bool) *Handler {
	return &Handler{auth: auth, log: log, secureCookie: secureCookie}
}

// Register mounts the routes on r.
func (h *Handler) Register(r gin.IRouter) {
	group := r.Group("/auth")
	group.POST("/login", h.login)
	group.POST("/logout", h.logout)
	group.GET("/me", h.me)
}

type loginRequest struct {
	Password string `json:"password" binding:"required"`
}

func (h *Handler) login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.Error(c, http.StatusBadRequest, "Password is required")
		return
	}

	token, err := h.auth.Login(req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			httpserver.Error(c, http.StatusUnauthorized, "Invalid password")
			return
		}
		h.log.ErrorContext(c.Request.Context(), "login failed", slog.Any("error", err))
		httpserver.Error(c, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.setCookie(c, token.Value, int(time.Until(token.ExpiresAt).Seconds()))
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Handler) logout(c *gin.Context) {
	h.setCookie(c, "", -1)
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Handler) me(c *gin.Context) {
	token, _ := c.Cookie(authtoken.CookieName)
	session, err := h.auth.Authenticate(token)
	if err != nil {
		httpserver.Error(c, http.StatusUnauthorized, "Not authenticated")
		return
	}
	c.JSON(http.StatusOK, gin.H{"authenticated": true, "role": session.Role})
}

func (h *Handler) setCookie(c *gin.Context, value string, maxAge int) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     authtoken.CookieName,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   h.secureCookie,
		SameSite: http.SameSiteLaxMode,
	})
}
