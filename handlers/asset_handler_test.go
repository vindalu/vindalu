package handlers

import (
	//"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	testResourcePath = "/v3/server/testserver"
)

/*
func Test_AssetWriteRequestHandler(t *testing.T) {

	buff := bytes.NewBuffer([]byte(`{"status":"enabled"}`))

	r, _ := http.NewRequest("GET", testResourcePath, buff)
	w := httptest.NewRecorder()

	testInv.AssetWriteRequestHandler(w, r)

	if w.Code != 200 {
		t.Fatalf("%#v\n", w)
	}
}
*/

func Test_AssetVersionsHandler(t *testing.T) {
	r, _ := http.NewRequest("GET", testResourcePath, nil)
	w := httptest.NewRecorder()

	testInv.AssetVersionsHandler(w, r)

	if w.Code != 200 {
		t.Fatalf("%v\n", w)
	}
}
