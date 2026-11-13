package httpapi

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPrivacyHasConfirmedIdentityAndEscapesOperator(t *testing.T) {
	router := (Handler{Operator: "ИП <Ermek>", SupportEmail: "a.ermek.kg@gmail.com"}).Router()
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest("GET", "/privacy", nil))
	if response.Code != 200 {
		t.Fatalf("privacy HTTP %d", response.Code)
	}
	body := response.Body.String()
	for _, expected := range []string{"ИП &lt;Ermek&gt;", "a.ermek.kg@gmail.com", "Кыргызская Республика", "16 дней", "12 месяцев"} {
		if !strings.Contains(body, expected) {
			t.Fatalf("missing policy content: %s", expected)
		}
	}
	if strings.Contains(body, "ИП <Ermek>") {
		t.Fatal("operator must be HTML escaped")
	}
}
func TestPrivacyDoesNotPublishMissingIdentity(t *testing.T) {
	response := httptest.NewRecorder()
	(Handler{SupportEmail: "a.ermek.kg@gmail.com"}).Router().ServeHTTP(response, httptest.NewRequest("GET", "/privacy", nil))
	if response.Code != 503 {
		t.Fatalf("missing operator must block publication, got %d", response.Code)
	}
}
