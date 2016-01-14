package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	testResourcePath    = "/v3/server/testserver"
	testResourcePathErr = "/v3/server/foo"
)

func Test_AssetVersionsHandler(t *testing.T) {
	r, _ := http.NewRequest("GET", testResourcePath, nil)
	w := httptest.NewRecorder()

	testInv.AssetVersionsHandler(w, r)

	if w.Code != 200 {
		t.Fatalf("%v\n", w)
	}
}

func Test_AssetOptionsHandler(t *testing.T) {
	r, _ := http.NewRequest("OPTIONS", "/v3/server", nil)
	w := httptest.NewRecorder()
	testInv.AssetOptionsHandler(w, r)

	if w.Code != 200 {
		t.Fatalf("%v\n", w)
	}
}

func Test_AssetVersionsOptionsHandler(t *testing.T) {
	r, _ := http.NewRequest("OPTIONS", "/v3/server/id/versions", nil)
	w := httptest.NewRecorder()
	testInv.AssetVersionsOptionsHandler(w, r)

	if w.Code != 200 {
		t.Fatalf("%v\n", w)
	}
}
