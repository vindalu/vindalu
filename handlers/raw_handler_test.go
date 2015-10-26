package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_Inventory_ESSRawHandler(t *testing.T) {
	r, _ := http.NewRequest("GET", "/v3/raw", nil)
	w := httptest.NewRecorder()

	testInv.ESSRawHandler(w, r)

	if w.Code != 200 {
		t.Fatalf("Failed: %v", w)
	}
	t.Log(w.Body.String())
}
