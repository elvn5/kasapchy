package httpapi

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/your-org/app/marketplace/internal/domain"
	"github.com/your-org/app/marketplace/internal/usecase"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Handler struct {
	Readiness      func(context.Context) error
	TrustedProxies []string
	Operator       string
	SupportEmail   string
	Service        *usecase.Service
	Log            *slog.Logger
}

func (h Handler) fail(c *gin.Context, e error) {
	code, message := http.StatusInternalServerError, "Сервис временно недоступен. Повторите позже."
	switch {
	case errors.Is(e, domain.ErrInvalid):
		code, message = 400, "Проверьте введённые данные."
	case errors.Is(e, domain.ErrUnauthorized):
		code, message = 401, "Не удалось войти. Проверьте данные и выбранную роль."
	case errors.Is(e, domain.ErrForbidden):
		code, message = 403, "Это действие недоступно для вашей роли."
	case errors.Is(e, domain.ErrNotFound):
		code, message = 404, "Запись не найдена."
	case errors.Is(e, domain.ErrConflict):
		code, message = 409, "Данные изменились, товара недостаточно или аккаунт имеет другую роль. Обновите экран."
	case errors.Is(e, domain.ErrRateLimit):
		code, message = 429, "Слишком много запросов. Попробуйте позже."
	}
	if code == 500 {
		h.Log.Error("request failed", "route", c.FullPath(), "error", e)
	}
	c.AbortWithStatusJSON(code, gin.H{"error": message})
}
func bind(c *gin.Context, v any) bool {
	if c.ShouldBindJSON(v) != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "Некорректный запрос."})
		return false
	}
	return true
}
func user(c *gin.Context) domain.User { return c.MustGet("user").(domain.User) }
func (h Handler) Router() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	if err := r.SetTrustedProxies(h.TrustedProxies); err != nil {
		panic("invalid trusted proxy configuration")
	}
	r.Use(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 32<<10)
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Cache-Control", "no-store")
		c.Next()
	})
	r.GET("/healthz", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	r.GET("/readyz", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if h.Readiness == nil || h.Readiness(ctx) != nil {
			c.JSON(503, gin.H{"status": "unavailable"})
			return
		}
		c.JSON(200, gin.H{"status": "ready"})
	})
	h.web(r)
	api := r.Group("/v1")
	for _, action := range []string{"register", "login"} {
		register := action == "register"
		api.POST("/auth/password/"+action, func(c *gin.Context) {
			var req struct{ Email, Password, Role string }
			if !bind(c, &req) {
				return
			}
			session, err := h.Service.PasswordAuth(c, req.Email, req.Password, req.Role, c.ClientIP(), register)
			if err != nil {
				h.fail(c, err)
				return
			}
			c.JSON(200, session)
		})
	}
	api.POST("/auth/email/request", func(c *gin.Context) {
		var req struct{ Email, Role string }
		if !bind(c, &req) {
			return
		}
		id, e := h.Service.RequestCode(c, req.Email, req.Role, c.ClientIP())
		if e != nil {
			h.fail(c, e)
			return
		}
		c.JSON(202, gin.H{"challengeId": id})
	})
	api.POST("/auth/email/verify", func(c *gin.Context) {
		var req struct {
			ChallengeID string `json:"challengeId"`
			Code        string
		}
		if !bind(c, &req) {
			return
		}
		session, e := h.Service.VerifyCode(c, req.ChallengeID, req.Code)
		if e != nil {
			h.fail(c, e)
			return
		}
		c.JSON(200, session)
	})
	api.POST("/auth/oidc/challenge", func(c *gin.Context) {
		var req struct {
			Provider, Role string
			ProofHash      string `json:"proofHash"`
		}
		if !bind(c, &req) {
			return
		}
		ok, e := h.Service.Repo.Allow(c, "oidc:"+c.ClientIP(), 30, time.Hour)
		if e != nil {
			h.fail(c, e)
			return
		}
		if !ok {
			h.fail(c, domain.ErrRateLimit)
			return
		}
		id, nonce, e := h.Service.StartOIDC(c, req.Provider, req.Role, req.ProofHash)
		if e != nil {
			h.fail(c, e)
			return
		}
		c.JSON(200, gin.H{"challengeId": id, "nonce": nonce})
	})
	api.POST("/auth/oidc/verify", func(c *gin.Context) {
		var req struct {
			ChallengeID string `json:"challengeId"`
			Provider    string
			IDToken     string `json:"idToken"`
		}
		if !bind(c, &req) {
			return
		}
		session, e := h.Service.FinishOIDC(c, req.ChallengeID, req.Provider, req.IDToken)
		if e != nil {
			h.fail(c, e)
			return
		}
		c.JSON(200, session)
	})
	api.POST("/auth/apple/callback", func(c *gin.Context) {
		if e := h.Service.AppleCallback(c, c.PostForm("state"), c.PostForm("id_token")); e != nil {
			h.fail(c, e)
			return
		}
		c.Header("Content-Security-Policy", "default-src 'none'")
		c.Data(200, "text/html; charset=utf-8", []byte("<!doctype html><meta charset=utf-8><title>Касапчы</title><h1>Вход подтверждён</h1><p>Вернитесь в приложение Касапчы.</p>"))
	})
	api.POST("/auth/apple/poll", func(c *gin.Context) {
		var req struct {
			ChallengeID string `json:"challengeId"`
			Proof       string
		}
		if !bind(c, &req) {
			return
		}
		session, e := h.Service.ApplePoll(c, req.ChallengeID, req.Proof)
		if errors.Is(e, domain.ErrNotFound) {
			c.Status(202)
			return
		}
		if e != nil {
			h.fail(c, e)
			return
		}
		c.JSON(200, session)
	})
	api.GET("/products", func(c *gin.Context) {
		var items []domain.Product
		var e error
		if auth := c.GetHeader("Authorization"); auth != "" {
			if !strings.HasPrefix(auth, "Bearer ") || len(auth) != 71 {
				h.fail(c, domain.ErrUnauthorized)
				return
			}
			u, err := h.Service.Repo.Authenticate(c, usecase.Hash(strings.TrimPrefix(auth, "Bearer ")))
			if err != nil {
				h.fail(c, err)
				return
			}
			items, e = h.Service.Repo.ProductsForUser(c, u.ID)
		} else {
			items, e = h.Service.Repo.Products(c)
		}
		if e != nil {
			h.fail(c, e)
			return
		}
		c.JSON(200, items)
	})
	protected := api.Group("")
	protected.Use(func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") || len(auth) != 71 {
			h.fail(c, domain.ErrUnauthorized)
			return
		}
		hash := usecase.Hash(strings.TrimPrefix(auth, "Bearer "))
		u, e := h.Service.Repo.Authenticate(c, hash)
		if e != nil {
			h.fail(c, e)
			return
		}
		c.Set("user", u)
		c.Set("tokenHash", hash)
		c.Next()
	})
	protected.POST("/products/:id/report", func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		var req struct{ Reason string }
		if err != nil || id < 1 {
			h.fail(c, domain.ErrInvalid)
			return
		}
		if !bind(c, &req) {
			return
		}
		req.Reason = strings.TrimSpace(req.Reason)
		if len([]rune(req.Reason)) < 3 || len([]rune(req.Reason)) > 500 {
			h.fail(c, domain.ErrInvalid)
			return
		}
		ok, err := h.Service.Repo.Allow(c, "report:"+user(c).ID, 10, 24*time.Hour)
		if err != nil {
			h.fail(c, err)
			return
		}
		if !ok {
			h.fail(c, domain.ErrRateLimit)
			return
		}
		if err = h.Service.Repo.ReportProduct(c, user(c).ID, id, req.Reason); err != nil {
			h.fail(c, err)
			return
		}
		c.Status(204)
	})
	protected.POST("/me/blocked-sellers", func(c *gin.Context) {
		var req struct {
			SellerID string `json:"sellerId"`
		}
		if !bind(c, &req) {
			return
		}
		if len(req.SellerID) != 36 {
			h.fail(c, domain.ErrInvalid)
			return
		}
		if err := h.Service.Repo.BlockSeller(c, user(c).ID, req.SellerID); err != nil {
			h.fail(c, err)
			return
		}
		c.Status(204)
	})
	protected.DELETE("/me/blocked-sellers", func(c *gin.Context) {
		if err := h.Service.Repo.UnblockSellers(c, user(c).ID); err != nil {
			h.fail(c, err)
			return
		}
		c.Status(204)
	})
	protected.GET("/me", func(c *gin.Context) { c.JSON(200, user(c)) })
	protected.PATCH("/me", func(c *gin.Context) {
		var req struct{ Name, Phone string }
		if !bind(c, &req) {
			return
		}
		req.Name = strings.TrimSpace(req.Name)
		req.Phone = strings.TrimSpace(req.Phone)
		if !domain.ValidProfile(req.Name, req.Phone) {
			h.fail(c, domain.ErrInvalid)
			return
		}
		u, e := h.Service.Repo.SaveProfile(c, user(c).ID, req.Name, req.Phone)
		if e != nil {
			h.fail(c, e)
			return
		}
		c.JSON(200, u)
	})
	protected.POST("/auth/logout", func(c *gin.Context) {
		if e := h.Service.Repo.Logout(c, c.GetString("tokenHash")); e != nil {
			h.fail(c, e)
			return
		}
		c.Status(204)
	})
	protected.DELETE("/me", func(c *gin.Context) {
		if e := h.Service.Repo.DeleteAccount(c, user(c).ID, c.GetString("tokenHash")); e != nil {
			h.fail(c, e)
			return
		}
		c.Status(204)
	})
	protected.POST("/products", func(c *gin.Context) {
		u := user(c)
		if u.Role != "Seller" {
			h.fail(c, domain.ErrForbidden)
			return
		}
		var req domain.NewProduct
		if !bind(c, &req) {
			return
		}
		if !req.Valid() {
			h.fail(c, domain.ErrInvalid)
			return
		}
		p, e := h.Service.Repo.Publish(c, u, req)
		if e != nil {
			h.fail(c, e)
			return
		}
		c.JSON(201, p)
	})
	protected.GET("/orders", func(c *gin.Context) {
		items, e := h.Service.Repo.Orders(c, user(c))
		if e != nil {
			h.fail(c, e)
			return
		}
		c.JSON(200, items)
	})
	protected.POST("/orders", func(c *gin.Context) {
		u := user(c)
		if u.Role != "Buyer" {
			h.fail(c, domain.ErrForbidden)
			return
		}
		var req domain.NewOrder
		if !bind(c, &req) {
			return
		}
		if !req.Valid() {
			h.fail(c, domain.ErrInvalid)
			return
		}
		if req.Delivery == "Pickup" {
			req.Address = ""
		}
		o, e := h.Service.Repo.Place(c, u, req)
		if e != nil {
			h.fail(c, e)
			return
		}
		c.JSON(201, o)
	})
	protected.POST("/orders/:id/advance", func(c *gin.Context) {
		u := user(c)
		if u.Role != "Seller" {
			h.fail(c, domain.ErrForbidden)
			return
		}
		id, e := strconv.ParseInt(c.Param("id"), 10, 64)
		if e != nil {
			h.fail(c, domain.ErrInvalid)
			return
		}
		var req struct {
			ExpectedStatus string `json:"expectedStatus"`
		}
		if !bind(c, &req) {
			return
		}
		o, e := h.Service.Repo.Advance(c, u, id, req.ExpectedStatus)
		if e != nil {
			h.fail(c, e)
			return
		}
		c.JSON(200, o)
	})
	return r
}
