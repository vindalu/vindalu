package handlers

import (
	"bytes"
	"net/http"
	"testing"
)

func Test_normalizeAssetType(t *testing.T) {
	if normalizeAssetType("ABC") != "abc" {
		t.Fatalf("Normalization failed")
	}
	if normalizeAssetType("") != "" {
		t.Fatalf("Normalization failed")
	}
}

func Test_parseRequestBody(t *testing.T) {
	// case 1: handle request body
	jsonStr := []byte(`{"os":"xenserver"}`)
	r1, _ := http.NewRequest("GET", "http://localhost:5454/v1/pool", bytes.NewBuffer(jsonStr))
	r1.Header.Set("X-Custom-Header", "myvalue")
	r1.Header.Set("Content-Type", "application/json")
	req, err1 := parseRequestBody(r1)
	if err1 != nil || req["os"] != "xenserver" {
		t.Fatalf("Error while parsing request Body!")
	}

	// case2 : ignore query params
	r2, _ := http.NewRequest("GET", "http://localhost:5454/v1/pool?os=xenserver", nil)
	if _, err2 := parseRequestBody(r2); err2 != nil {
		t.Fatalf("Error while parsing request query params!")
	}
}

func Test_getQueryParamsFromRequest(t *testing.T) {
	r, _ := http.NewRequest("GET", "http://localhost:5454/v1/pool?os=xenserver&os=ubuntu", nil)
	req, err := getQueryParamsFromRequest(r)
	if err != nil || req["os"] != "xenserver|ubuntu" {
		t.Fatalf("Error while parsing request query params!")
	}
}
