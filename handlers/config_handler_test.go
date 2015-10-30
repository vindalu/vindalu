package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_Inventory_ConfigHandler(t *testing.T) {
	r, _ := http.NewRequest("GET", "/config", nil)
	w := httptest.NewRecorder()

	testInv.ConfigHttpHandler(w, r)

	if w.Code != 200 {
		t.Fatalf("Failed: %v", w)
	}
	t.Log(w.Body.String())
}
