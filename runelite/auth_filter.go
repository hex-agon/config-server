package runelite

import (
	"context"
	"database/sql"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

const authHeader = "Runelite-Auth"
const ctxToken = "authToken"

type AuthFilter struct {
	repository SessionRepository
}

func NewAuthFilter(repository SessionRepository) AuthFilter {
	return AuthFilter{
		repository: repository,
	}
}

func (a *AuthFilter) Filtered(
	handler AuthorizedHttpHandle,
) httprouter.Handle {
	return func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		authToken := request.Header.Get(authHeader)

		if authToken == "" {
			http.Error(writer, "Unauthorized", http.StatusUnauthorized)
		} else {
			userId, err := a.repository.FindUserIdByUuid(authToken)

			if err == sql.ErrNoRows {
				http.Error(writer, "Unauthorized", http.StatusUnauthorized)
			} else if err != nil {
				http.Error(writer, "Internal server error", http.StatusInternalServerError)
			} else {
				newCtx := context.WithValue(request.Context(), ctxToken, authToken)
				handler(userId, writer, request.WithContext(newCtx), params)
			}
		}
	}
}
