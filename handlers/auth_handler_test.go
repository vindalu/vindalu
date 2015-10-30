package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/context"
)

func Test_Inventory_AuthTokenHandler(t *testing.T) {
	r, _ := http.NewRequest("GET", "/auth/access_token", nil)

	context.Set(r, Username, "admin")
	context.Set(r, IsAdmin, true)

	w := httptest.NewRecorder()

	testInv.AuthTokenHandler(w, r)

	if w.Code != 200 {
		t.Fatalf("Failed: %v", w)
	}
	t.Log(w.Body.String())
}
