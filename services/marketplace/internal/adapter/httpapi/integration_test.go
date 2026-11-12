package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/your-org/app/marketplace/internal/adapter/postgres"
	"github.com/your-org/app/marketplace/internal/domain"
	"github.com/your-org/app/marketplace/internal/usecase"
	"github.com/your-org/app/marketplace/migrations"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"
)

type captureMail struct {
	mu    sync.Mutex
	codes map[string]string
}

func (m *captureMail) SendCode(_ context.Context, email, code string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.codes[email] = code
	return nil
}

type fixture struct {
	t      *testing.T
	server *httptest.Server
	db     *pgxpool.Pool
	mail   *captureMail
}

func setup(t *testing.T) *fixture {
	t.Helper()
	dsn := os.Getenv("MARKET_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set MARKET_TEST_DATABASE_URL to a disposable PostgreSQL database")
	}
	ctx := context.Background()
	admin, e := pgxpool.New(ctx, dsn)
	if e != nil {
		t.Fatal(e)
	}
	schema := fmt.Sprintf("test_%d", time.Now().UnixNano())
	if _, e = admin.Exec(ctx, "CREATE SCHEMA "+schema); e != nil {
		t.Fatal(e)
	}
	config, e := pgxpool.ParseConfig(dsn)
	if e != nil {
		t.Fatal(e)
	}
	config.ConnConfig.RuntimeParams["search_path"] = schema
	db, e := pgxpool.NewWithConfig(ctx, config)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = db.Exec(ctx, migrations.Initial); e != nil {
		t.Fatal(e)
	}
	mail := &captureMail{codes: map[string]string{}}
	service := &usecase.Service{Repo: &postgres.Repository{DB: db}, Mail: mail, Secret: []byte("integration-test-secret-not-production")}
	server := httptest.NewServer((Handler{Service: service, Log: slog.New(slog.NewTextHandler(io.Discard, nil))}).Router())
	t.Cleanup(func() {
		server.Close()
		db.Close()
		_, _ = admin.Exec(ctx, "DROP SCHEMA "+schema+" CASCADE")
		admin.Close()
	})
	return &fixture{t, server, db, mail}
}
func (f *fixture) request(method, path, token string, body any, want int, out any) {
	f.t.Helper()
	payload, _ := json.Marshal(body)
	req, _ := http.NewRequest(method, f.server.URL+"/v1"+path, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, e := http.DefaultClient.Do(req)
	if e != nil {
		f.t.Fatal(e)
	}
	defer res.Body.Close()
	if res.StatusCode != want {
		b, _ := io.ReadAll(res.Body)
		f.t.Fatalf("%s %s: want %d got %d %s", method, path, want, res.StatusCode, b)
	}
	if out != nil {
		if e = json.NewDecoder(res.Body).Decode(out); e != nil {
			f.t.Fatal(e)
		}
	}
}
func (f *fixture) login(email, role string) domain.Session {
	f.t.Helper()
	var challenge map[string]string
	f.request("POST", "/auth/email/request", "", map[string]string{"email": email, "role": role}, 202, &challenge)
	f.mail.mu.Lock()
	code := f.mail.codes[email]
	f.mail.mu.Unlock()
	if len(code) != 6 || code == "1234" {
		f.t.Fatal("invalid generated code")
	}
	var session domain.Session
	f.request("POST", "/auth/email/verify", "", map[string]string{"challengeId": challenge["challengeId"], "code": code}, 200, &session)
	f.request("POST", "/auth/email/verify", "", map[string]string{"challengeId": challenge["challengeId"], "code": code}, 401, nil)
	return session
}
func TestMarketplaceIsolationAndIdempotency(t *testing.T) {
	f := setup(t)
	seller := f.login("seller@example.com", "Seller")
	buyer := f.login("buyer@example.com", "Buyer")
	other := f.login("other@example.com", "Seller")
	var product domain.Product
	input := domain.NewProduct{Name: "Говяжья мякоть", Category: "Beef", Price: 680, Stock: 2}
	f.request("POST", "/products", buyer.Token, input, 403, nil)
	f.request("POST", "/products", seller.Token, input, 201, &product)
	orderInput := domain.NewOrder{ProductID: product.ID, Quantity: 2, Delivery: "Courier", Address: "Чуй 100", RequestID: "stable-request-id-001"}
	var order domain.Order
	f.request("POST", "/orders", buyer.Token, orderInput, 201, &order)
	if order.Total != 1510 {
		t.Fatalf("incorrect server total: %d", order.Total)
	}
	var again domain.Order
	f.request("POST", "/orders", buyer.Token, orderInput, 201, &again)
	if again.ID != order.ID {
		t.Fatal("duplicate order")
	}
	orderInput.RequestID = "another-request-id-002"
	f.request("POST", "/orders", buyer.Token, orderInput, 409, nil)
	var orders []domain.Order
	f.request("GET", "/orders", other.Token, nil, 200, &orders)
	if len(orders) != 0 {
		t.Fatal("cross-seller disclosure")
	}
	path := fmt.Sprintf("/orders/%d/advance", order.ID)
	f.request("POST", path, other.Token, map[string]string{"expectedStatus": "Pending"}, 409, nil)
	f.request("POST", path, buyer.Token, map[string]string{"expectedStatus": "Pending"}, 403, nil)
	f.request("POST", path, seller.Token, map[string]string{"expectedStatus": "Pending"}, 200, &again)
	if again.Status != "Accepted" {
		t.Fatal("transition not applied")
	}
	f.request("POST", path, seller.Token, map[string]string{"expectedStatus": "Pending"}, 409, nil)
	f.request("POST", "/auth/logout", buyer.Token, nil, 204, nil)
	f.request("GET", "/me", buyer.Token, nil, 401, nil)
}
func TestProfileDeletionAndOptionalPhone(t *testing.T) {
	f := setup(t)
	session := f.login("buyer@example.com", "Buyer")
	var u domain.User
	f.request("PATCH", "/me", session.Token, map[string]string{"name": "", "phone": ""}, 200, &u)
	f.request("PATCH", "/me", session.Token, map[string]string{"name": "", "phone": "letters"}, 400, nil)
	f.request("PATCH", "/me", session.Token, map[string]string{"name": "", "phone": "+996 700 123 456"}, 200, &u)
	f.request("DELETE", "/me", session.Token, nil, 204, nil)
	f.request("GET", "/me", session.Token, nil, 401, nil)
	var count int
	if e := f.db.QueryRow(context.Background(), "SELECT count(*) FROM market_users").Scan(&count); e != nil || count != 0 {
		t.Fatal("account not deleted", e)
	}
}
func TestOtpAttemptsCooldownAndRoleCannotBeChanged(t *testing.T) {
	f := setup(t)
	var c map[string]string
	request := map[string]string{"email": "locked@example.com", "role": "Buyer"}
	f.request("POST", "/auth/email/request", "", request, 202, &c)
	f.request("POST", "/auth/email/request", "", request, 429, nil)
	f.mail.mu.Lock()
	real := f.mail.codes[request["email"]]
	f.mail.mu.Unlock()
	wrong := "000000"
	if real == wrong {
		wrong = "111111"
	}
	for i := 0; i < 5; i++ {
		f.request("POST", "/auth/email/verify", "", map[string]string{"challengeId": c["challengeId"], "code": wrong}, 401, nil)
	}
	f.request("POST", "/auth/email/verify", "", map[string]string{"challengeId": c["challengeId"], "code": real}, 401, nil)
	session := f.login("role@example.com", "Buyer")
	_, e := f.db.Exec(context.Background(), "DELETE FROM market_limits")
	if e != nil {
		t.Fatal(e)
	}
	f.request("POST", "/auth/email/request", "", map[string]string{"email": session.User.Email, "role": "Seller"}, 202, &c)
	f.mail.mu.Lock()
	real = f.mail.codes[session.User.Email]
	f.mail.mu.Unlock()
	f.request("POST", "/auth/email/verify", "", map[string]string{"challengeId": c["challengeId"], "code": real}, 409, nil)
}
