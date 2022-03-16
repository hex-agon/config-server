package runelite

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireAuthMissingToken(t *testing.T) {
	nextHandler := func(userId int, writer http.ResponseWriter, request *http.Request, params httprouter.Params) {}
	middleware := Authenticated(nextHandler)

	recorder := httptest.NewRecorder()
	request, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	middleware(recorder, request, nil)

	if status := recorder.Code; status != http.StatusUnauthorized {
		t.Errorf("Invalid http got status %d but expected %d", status, http.StatusUnauthorized)
	}
}

func TestRequireAuthToken(t *testing.T) {
	expectedUserId := 1000
	recorder := httptest.NewRecorder()

	nextHandler := func(userId int, writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		if userId != expectedUserId {
			t.Errorf("Unexpected token, got %s but expected %s, http status %d", userId, expectedUserId, recorder.Code)
		}
	}
	middleware := Authenticated(nextHandler)

	request, err := http.NewRequest("GET", "/test", nil)
	request.Header.Add(authHeader, "uuid")
	if err != nil {
		t.Fatal(err)
	}
	middleware(recorder, request, nil)
}
