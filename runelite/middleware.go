package runelite

import (
	"context"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

const authHeader = "Runelite-Auth"
const ctxToken = "authToken"

func Authenticated(
	handler AuthorizedHttpHandle,
) httprouter.Handle {
	return func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		authToken := request.Header.Get(authHeader)

		if authToken == "" {
			http.Error(writer, "Unauthorized", http.StatusUnauthorized)
		} else {
			newCtx := context.WithValue(request.Context(), ctxToken, authToken)
			// TODO: Fetch userId from somewhere
			handler(1000, writer, request.WithContext(newCtx), params)
		}
	}
}
