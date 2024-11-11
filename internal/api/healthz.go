package api

import (
	"net/http"
)

type HealthzResponse struct {
	Status  string `json:"status"`
	Env     string `json:"env"`
	Version string `json:"version"`
}

// Health Check godoc
//
//	@Tags		health
//	@Produce	json
//	@Success	200	{object}	HealthzResponse	"Service is running"
//	@Router		/healthz [get]
func (app *Application) handlerHealthz(w http.ResponseWriter, r *http.Request) {
	res := HealthzResponse{
		Status:  "ok",
		Env:     app.Config.Environment,
		Version: app.Config.Version,
	}

	respondWithJSON(w, http.StatusOK, res)
}
