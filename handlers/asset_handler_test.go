package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	testResourcePath = "/v3/server/testserver"
)

func Test_AssetVersionsHandler(t *testing.T) {
	r, _ := http.NewRequest("GET", testResourcePath, nil)
	w := httptest.NewRecorder()

	testInv.AssetVersionsHandler(w, r)

	if w.Code != 200 {
		t.Fatalf("%v\n", w)
	}
}
