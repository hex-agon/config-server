package main

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
)

type AuthorizedHttpHandle func(userId int64, writer http.ResponseWriter, request *http.Request, params httprouter.Params)

type Handlers struct {
	logger     *zap.Logger
	repository ConfigRepository
}

func NewHandlers(logger *zap.Logger, repository ConfigRepository) *Handlers {
	return &Handlers{
		logger:     logger,
		repository: repository,
	}
}

func (h *Handlers) HandleGet(userId int64, writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	configuration, err := h.repository.FindByUserId(request.Context(), userId)

	if configuration == nil {
		if err == nil {
			http.NotFound(writer, request)
			return
		} else {
			http.Error(writer, "Internal server error", http.StatusInternalServerError)
			h.logger.Error("Error fetching config document", zap.Error(err))
			return
		}
	} else {
		err = json.NewEncoder(writer).Encode(configuration)

		if err != nil {
			http.Error(writer, "Internal server error", http.StatusInternalServerError)
			h.logger.Error("Error serializing config json", zap.Error(err))
		}
	}
}

func (h *Handlers) HandlePut(userId int64, writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	key := params.ByName("key")
	value, err := ioutil.ReadAll(request.Body)

	if err != nil {
		http.Error(writer, "Invalid request body", http.StatusBadRequest)
		h.logger.Error("Failed to read request body", zap.Error(err))
		return
	}
	err = h.repository.Save(request.Context(), userId, &ConfigEntry{
		Key:   key,
		Value: string(value),
	})

	if err != nil {
		http.Error(writer, "Update failed", http.StatusInternalServerError)
		h.logger.Error("Failed to update config entry", zap.Error(err))
	}
}

func (h *Handlers) HandlePatch(userId int64, writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	var configuration Configuration
	err := json.NewDecoder(request.Body).Decode(&configuration)

	if err != nil {
		http.Error(writer, "Invalid request body", http.StatusBadRequest)
		h.logger.Error("Error decoding configuration json", zap.Error(err))
		return
	}
	failedKeys, err := h.repository.SaveBatch(request.Context(), userId, &configuration)

	if err != nil {
		http.Error(writer, "Update failed", http.StatusInternalServerError)
		h.logger.Error("Failed to batch update config entries", zap.Error(err))
	}
	err = json.NewEncoder(writer).Encode(failedKeys)

	if err != nil {
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		h.logger.Error("Error serializing response", zap.Error(err))
	}
}

func (h *Handlers) HandleDelete(userId int64, writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	key := params.ByName("key")
	err := h.repository.DeleteKey(request.Context(), userId, key)

	if err != nil {
		http.Error(writer, "Delete failed", http.StatusInternalServerError)
		h.logger.Error("Error deleting config entry", zap.Error(err))
	}
}
