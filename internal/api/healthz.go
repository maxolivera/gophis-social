package api

import (
	"net/http"
)

func (app *Application) handlerHealthz(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Status  string `json:"status"`
		Env     string `json:"env"`
		Version string `json:"version"`
	}

	res := response{
		Status: "ok",
		Env: app.Config.Environment,
		Version: app.Config.Version,
	}

	respondWithJSON(w, http.StatusOK, res)
}
