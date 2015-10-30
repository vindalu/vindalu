package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_StatusHandler(t *testing.T) {

	r, _ := http.NewRequest("GET", "/status", nil)

	w := httptest.NewRecorder()

	testInv.StatusHandler(w, r)

	if w.Code != 200 {
		t.Fatalf("Failed: %v", w)
	}
	t.Log(w.Body.String())

}
