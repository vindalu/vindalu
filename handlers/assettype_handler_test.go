package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_Inventory_ListAssetTypesHandler(t *testing.T) {
	r, _ := http.NewRequest("GET", "/v3/virtualserver/properties", nil)

	w := httptest.NewRecorder()

	testInv.ListAssetTypesHandler(w, r)

	if w.Code != 200 {
		t.Fatalf("Failed: %v", w)
	}
	t.Log(w.Body.String())
}

func Test_Inventory_AssetTypeGetHandler(t *testing.T) {
	r, _ := http.NewRequest("GET", "/v3/virtualserver", nil)

	w := httptest.NewRecorder()

	testInv.AssetTypeGetHandler(w, r)

	if w.Code != 200 {
		t.Fatalf("Failed: %v", w)
	}
	t.Log(w.Body.String())
}

func Test_Inventory_AssetTypeListOptionsHandler(t *testing.T) {
	r, _ := http.NewRequest("GET", "/v3/", nil)

	w := httptest.NewRecorder()

	testInv.AssetTypeListOptionsHandler(w, r)

	if w.Code != 200 {
		t.Fatalf("Failed: %v", w)
	}
	t.Log(w.Body.String())
}

func Test_Inventory_AssetTypeOptionsHandler(t *testing.T) {
	r, _ := http.NewRequest("GET", "/v3/virtualserver", nil)

	w := httptest.NewRecorder()

	testInv.AssetTypeOptionsHandler(w, r)

	if w.Code != 200 {
		t.Fatalf("Failed: %v", w)
	}
	t.Log(w.Body.String())
}
