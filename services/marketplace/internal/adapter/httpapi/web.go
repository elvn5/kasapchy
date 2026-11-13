package httpapi

import (
	"embed"
	"github.com/gin-gonic/gin"
	"html/template"
)

//go:embed web/*
var pages embed.FS

func (h Handler) web(r *gin.Engine) {
	for path, file := range map[string]string{"/content-rules": "web/content-rules.html", "/delete-account": "web/delete.html", "/account-deletion.js": "web/delete.js", "/account-deletion.css": "web/delete.css"} {
		content, _ := pages.ReadFile(file)
		mime := "text/html; charset=utf-8"
		if path == "/account-deletion.js" {
			mime = "text/javascript; charset=utf-8"
		}
		if path == "/account-deletion.css" {
			mime = "text/css; charset=utf-8"
		}
		r.GET(path, func(c *gin.Context) {
			c.Header("Content-Security-Policy", "default-src 'none'; script-src 'self'; style-src 'self'; connect-src 'self'; base-uri 'none'; frame-ancestors 'none'; form-action 'self'")
			c.Header("Referrer-Policy", "no-referrer")
			c.Data(200, mime, content)
		})
	}
	r.GET("/privacy", func(c *gin.Context) {
		if h.Operator == "" || h.SupportEmail == "" {
			c.String(503, "Политика конфиденциальности готовится. Публичный запуск ещё не состоялся.")
			return
		}
		c.Header("Content-Security-Policy", "default-src 'none'; style-src 'self'; frame-ancestors 'none'")
		c.Header("Content-Type", "text/html; charset=utf-8")
		_ = privacy.Execute(c.Writer, map[string]string{"Operator": h.Operator, "Support": h.SupportEmail})
	})
}

var privacy = template.Must(template.ParseFS(pages, "web/privacy.html"))
