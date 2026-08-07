package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Ghost-Unison/Chirpy/internal/auth"
	"github.com/google/uuid"
)

type polkaWebhooksRequest struct {
	Event string                   `json:"event"`
	Data  polkaWebhooksRequestData `json:"data"`
}

type polkaWebhooksRequestData struct {
	UserID uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) polkaWebhooksHandler(w http.ResponseWriter, r *http.Request) {

	// validate request
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if apiKey != cfg.polkaKey {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// parse request body
	var req polkaWebhooksRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//check event
	if req.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	//查询，如果用户不存在返回404，然后更新，更新失败返回500 - 可以合并
	/*
		 _, err := cfg.dbQueries.GetUserByID(r.Context(), req.Data.UserID)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, err = cfg.dbQueries.UpdateUserToChirpyRedByID(r.Context(), req.Data.UserID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	*/
	//upgrade user to chirpy red
	_, err = cfg.dbQueries.UpdateUserToChirpyRedByID(r.Context(), req.Data.UserID)
	if err != nil {
		//当没有匹配行时它本身就会返回 sql.ErrNoRows 这时返回404
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
