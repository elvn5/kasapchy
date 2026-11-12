package domain

import "testing"

func TestValidation(t *testing.T) {
	for _, s := range []string{"", "x", "a\r\n@example.com", "Name <a@example.com>"} {
		if _, e := NormalizeEmail(s); e == nil {
			t.Errorf("accepted %q", s)
		}
	}
	if s, e := NormalizeEmail(" A@Example.com "); e != nil || s != "a@example.com" {
		t.Fatal(s, e)
	}
	if !ValidProfile("", "") || !ValidProfile("", "+996 700 123 456") || ValidProfile("", "abcdefg") {
		t.Fatal("phone validation")
	}
	if NextStatus("Completed") != "" || NextStatus("Pending") != "Accepted" {
		t.Fatal("state machine")
	}
	if (NewProduct{Name: "OK", Category: "Beef", Price: 0, Stock: 1}).Valid() {
		t.Fatal("zero price")
	}
}
