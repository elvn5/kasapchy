package httpapi

import (
	"embed"
	"github.com/gin-gonic/gin"
	"html/template"
)

//go:embed web/*
var pages embed.FS

func (h Handler) web(r *gin.Engine) {
	for path, file := range map[string]string{"/delete-account": "web/delete.html", "/account-deletion.js": "web/delete.js", "/account-deletion.css": "web/delete.css"} {
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

var privacy = template.Must(template.New("privacy").Parse(`<!doctype html><html lang="ru"><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><link rel="stylesheet" href="/account-deletion.css"><title>Конфиденциальность — Касапчы</title><main><p class="brand">касапчы</p><h1>Политика конфиденциальности</h1><p>Оператор: {{.Operator}}. Поддержка: {{.Support}}.</p><h2>Какие данные используются</h2><p>Email нужен для входа и связи по заказу. Имя и телефон необязательны. При доставке используется адрес. Мы храним роль, товары продавца, состав и статусы заказов. Продавец видит email покупателя и адрес своего заказа; покупатель видит сведения о продавце и товаре.</p><h2>Вход через Google и Apple</h2><p>По вашему выбору сервис идентификации подтверждает email. Пароль Google или Apple приложение не получает. Токены проверяются сервером. В Android токен сессии хранится в зашифрованном виде с Android Keystore.</p><h2>Хранение и удаление</h2><p>Аккаунт и заказы хранятся до удаления. Удаление аккаунта удаляет профиль, сессии, товары и связанные заказы. Коды входа действуют 10 минут; просроченные коды и идентификаторы ограничения частоты удаляются фоновым процессом. Сессии действуют до 30 дней или до выхода. Сведения не используются для рекламы; аналитические и рекламные SDK не подключены.</p><p>Удалить аккаунт можно в профиле приложения или на <a href="/delete-account">странице удаления</a>, подтвердив email. Обращения по данным принимаются по указанному email поддержки.</p><p>Эта политика описывает текущую реализацию сервиса. Оператор отвечает за актуальность контактных данных, размещение серверов и применимые требования к хранению данных.</p></main></html>`))
