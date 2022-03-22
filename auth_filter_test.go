package main

import (
	"database/sql"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var filter *AuthFilter

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}

type mockSessionCache struct {
}

func (m mockSessionCache) GetUserId(uuid string) (int64, error) {
	if uuid == "missing" {
		return -1, sql.ErrNoRows
	}
	return 1000, nil
}

func setup() {
	filter = NewAuthFilter(mockSessionCache{})
}

func TestRequireAuthMissingToken(t *testing.T) {
	nextHandler := func(userId int64, writer http.ResponseWriter, request *http.Request, params httprouter.Params) {}
	filteredHandler := filter.Filtered(nextHandler)

	recorder := httptest.NewRecorder()
	request, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	filteredHandler(recorder, request, nil)

	if status := recorder.Code; status != http.StatusUnauthorized {
		t.Errorf("Invalid http got status %d but expected %d", status, http.StatusUnauthorized)
	}
}

func TestRequireAuthToken(t *testing.T) {
	expectedUserId := int64(1000)
	recorder := httptest.NewRecorder()

	nextHandler := func(userId int64, writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		if userId != 1000 {
			t.Errorf("Unexpected token, got %d but expected %d, http status %d", userId, expectedUserId, recorder.Code)
		}
	}
	filteredHandler := filter.Filtered(nextHandler)

	request, err := http.NewRequest("GET", "/test", nil)
	request.Header.Add(authHeader, "uuid")
	if err != nil {
		t.Fatal(err)
	}
	filteredHandler(recorder, request, nil)
}

func TestMissingSession(t *testing.T) {
	recorder := httptest.NewRecorder()

	nextHandler := func(userId int64, writer http.ResponseWriter, request *http.Request, params httprouter.Params) {}
	filteredHandler := filter.Filtered(nextHandler)

	request, err := http.NewRequest("GET", "/test", nil)
	request.Header.Add(authHeader, "missing")
	if err != nil {
		t.Fatal(err)
	}
	filteredHandler(recorder, request, nil)
	if status := recorder.Code; status != http.StatusUnauthorized {
		t.Errorf("Invalid http got status %d but expected %d", status, http.StatusUnauthorized)
	}
}
