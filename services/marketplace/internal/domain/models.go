package domain

import (
	"errors"
	"net/mail"
	"regexp"
	"strings"
	"time"
)

var (
	ErrInvalid      = errors.New("invalid input")
	ErrUnauthorized = errors.New("authentication required")
	ErrForbidden    = errors.New("forbidden")
	ErrNotFound     = errors.New("not found")
	ErrConflict     = errors.New("conflict")
	ErrRateLimit    = errors.New("too many requests")
)

type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Provider string `json:"provider"`
}
type Product struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Price       int    `json:"price"`
	Seller      string `json:"seller"`
	SellerID    string `json:"sellerId"`
	Stock       int    `json:"stock"`
	Description string `json:"description"`
}
type Order struct {
	ID         int64     `json:"id"`
	Product    Product   `json:"product"`
	Quantity   int       `json:"quantity"`
	BuyerEmail string    `json:"buyerEmail"`
	Delivery   string    `json:"delivery"`
	Address    string    `json:"address"`
	Status     string    `json:"status"`
	Total      int       `json:"total"`
	CreatedAt  time.Time `json:"createdAt"`
}
type Session struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
type NewProduct struct {
	Name        string `json:"name"`
	Category    string `json:"category"`
	Price       int    `json:"price"`
	Stock       int    `json:"stock"`
	Description string `json:"description"`
}
type NewOrder struct {
	ProductID int64  `json:"productId"`
	Quantity  int    `json:"quantity"`
	Delivery  string `json:"delivery"`
	Address   string `json:"address"`
	RequestID string `json:"requestId"`
}

func NormalizeEmail(s string) (string, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	a, e := mail.ParseAddress(s)
	if e != nil || a.Address != s || len(s) > 254 || !strings.Contains(s, ".") {
		return "", ErrInvalid
	}
	return s, nil
}
func ValidRole(s string) bool { return s == "Buyer" || s == "Seller" }
func ValidProfile(name, phone string) bool {
	if len([]rune(name)) > 60 || len(phone) > 30 {
		return false
	}
	if phone == "" {
		return true
	}
	if !regexp.MustCompile(`^\+?[0-9 ()-]+$`).MatchString(phone) {
		return false
	}
	digits := regexp.MustCompile(`[^0-9]`).ReplaceAllString(phone, "")
	return len(digits) >= 7 && len(digits) <= 15
}
func (p NewProduct) Valid() bool {
	return len([]rune(strings.TrimSpace(p.Name))) >= 2 && len([]rune(p.Name)) <= 60 && p.Price >= 1 && p.Price <= 100000 && p.Stock >= 1 && p.Stock <= 10000 && len([]rune(p.Description)) <= 2000 && (p.Category == "Beef" || p.Category == "Lamb" || p.Category == "Poultry")
}
func (o NewOrder) Valid() bool {
	return o.ProductID > 0 && o.Quantity >= 1 && o.Quantity <= 10000 && len(o.RequestID) >= 16 && len(o.RequestID) <= 100 && len([]rune(o.Address)) <= 500 && (o.Delivery == "Pickup" || (o.Delivery == "Courier" && len([]rune(strings.TrimSpace(o.Address))) >= 5))
}
func NextStatus(s string) string {
	switch s {
	case "Pending":
		return "Accepted"
	case "Accepted":
		return "Ready"
	case "Ready":
		return "Completed"
	}
	return ""
}
