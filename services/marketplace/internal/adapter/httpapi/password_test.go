package httpapi

import (
	"context"
	"github.com/your-org/app/marketplace/internal/domain"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"testing"
)

func TestPasswordRegistrationLoginAndIsolation(t *testing.T) {
	f := setup(t)
	credentials := map[string]string{"email": " NEW@example.com ", "password": "a-long-test-password", "role": "Buyer"}
	var original, again domain.Session
	f.request("POST", "/auth/password/register", "", credentials, 200, &original)
	if original.User.Email != "new@example.com" || original.Token == "" {
		t.Fatal("invalid session")
	}
	var stored string
	if err := f.db.QueryRow(context.Background(), "SELECT password_hash FROM market_users WHERE email=$1", original.User.Email).Scan(&stored); err != nil {
		t.Fatal(err)
	}
	if stored == credentials["password"] || bcrypt.CompareHashAndPassword([]byte(stored), []byte(credentials["password"])) != nil {
		t.Fatal("password not securely hashed")
	}
	f.request("POST", "/auth/password/login", "", credentials, 200, &again)
	if again.User.ID != original.User.ID || again.Token == original.Token {
		t.Fatal("login must issue fresh session for same account")
	}
	credentials["password"] = "wrong-test-password"
	f.request("POST", "/auth/password/login", "", credentials, 401, nil)
	f.request("POST", "/auth/password/register", "", credentials, 401, nil)
	credentials["password"] = "a-long-test-password"
	credentials["role"] = "Seller"
	f.request("POST", "/auth/password/login", "", credentials, 401, nil)
	f.request("GET", "/me", original.Token, nil, 200, nil)
	// Verified email possession must not silently attach to an unverified password identity.
	var challenge map[string]string
	f.request("POST", "/auth/email/request", "", map[string]string{"email": "new@example.com", "role": "Buyer"}, 202, &challenge)
	f.request("POST", "/auth/email/verify", "", map[string]string{"challengeId": challenge["challengeId"], "code": f.mail.codes["new@example.com"]}, 401, nil)
	f.request("DELETE", "/me", again.Token, nil, 204, nil)
	f.request("GET", "/me", original.Token, nil, 401, nil)
}

func TestPasswordValidationRateLimitAndExistingIdentity(t *testing.T) {
	f := setup(t)
	existing := f.login("verified@example.com", "Buyer")
	input := map[string]string{"email": existing.User.Email, "password": "a-long-test-password", "role": "Buyer"}
	f.request("POST", "/auth/password/register", "", input, 401, nil)
	// Knowing the dummy hash's input cannot authenticate an OTP-only identity.
	input["password"] = "invalid-account-dummy-credential"
	f.request("POST", "/auth/password/login", "", input, 401, nil)
	input["email"] = "missing@example.com"
	f.request("POST", "/auth/password/login", "", input, 401, nil)
	for _, p := range []string{"", "short", strings.Repeat("a", 73)} {
		input["password"] = p
		f.request("POST", "/auth/password/register", "", input, 400, nil)
	}
	input["email"] = "rate@example.com"
	input["password"] = "a-long-test-password"
	for i := 0; i < 10; i++ {
		f.request("POST", "/auth/password/login", "", input, 401, nil)
	}
	f.request("POST", "/auth/password/login", "", input, 429, nil)
	f.request("GET", "/me", existing.Token, nil, 200, nil)
}
