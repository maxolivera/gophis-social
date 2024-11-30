package api

import (
	"encoding/json"
	"net/http"
)

const MAX_BYTES = 1_048_578 // 1 MB

// TODO(maolivera): Better functions to return JSON respones, following some kind of standard
// TODO(maolivera): Respond with JSON should not respond with application/json if code is NoContent
// TODO(maolivera): Split function into status code

func (app *Application) respondWithJSON(w http.ResponseWriter, r *http.Request, code int, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		app.Logger.Errorf("Failed to marshal JSON response: %v", payload)
		app.Logger.Errorf("Error: %v", err)

		// NOTE(maolivera): Hardcoding to avoid adding http.Request parameter. Maybe adding and re-use function?

		w.Write([]byte("{\"error\": \"the server encountered a problem\" }"))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	app.Logger.Infow("sending HTTP response", "http_code", code, "method", r.Method, "path", r.URL.Path)

	if code != http.StatusNoContent {
		w.Header().Add("Content-Type", "application/json")
	}

	w.WriteHeader(code)

	if code != http.StatusNoContent {
		w.Write(data)
	}
}

// Will sent a response to the client with code `code` and message `message`, and will log the error `err`
func (app *Application) respondWithError(w http.ResponseWriter, r *http.Request, code int, err error, message string) {
	type response struct {
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	res := response{}

	res.Error.Code = code
	// default messages for specific codes
	if message == "" {
		switch code {
		case http.StatusInternalServerError:
			res.Error.Message = "the server encountered a problem"
		default:
			app.Logger.Warn("the HTTP code %d do not support default message, an empty message will be sent", code)
			res.Error.Message = ""
		}
	} else {
		res.Error.Message = message
	}
	// Log level based on error type
	switch code {
	case http.StatusInternalServerError:
		app.Logger.Errorw("there was an error", "http_code", code, "method", r.Method, "path", r.URL.Path, "error", err.Error())
	case http.StatusForbidden:
		app.Logger.Warnw("trying to access forbidden place", "http_code", code, "method", r.Method, "path", r.URL.Path, "error", err.Error())
	case http.StatusTooManyRequests:
		// Nothing as it is already logged
	default:
		app.Logger.Infow("responding with code > 4xx", "http_code", code, "method", r.Method, "path", r.URL.Path, "error", err.Error())
	}

	app.respondWithJSON(w, r, code, res)
}

func readJSON(w http.ResponseWriter, r *http.Request, data any) error {
	r.Body = http.MaxBytesReader(w, r.Body, MAX_BYTES)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	return decoder.Decode(data)
}

func (app *Application) unauthorizedBasicErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.Logger.Warnf("unauthorized basic error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	// NOTE: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/WWW-Authenticate
	w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)

	app.respondWithError(w, r, http.StatusUnauthorized, err, "Unauthorized")
}

func (app *Application) rateLimiteExceededErrorResponse(w http.ResponseWriter, r *http.Request, retryAfter string) {
	app.Logger.Warnw("rate limiter exceed", "method", r.Method, "path", r.URL.Path)

	w.Header().Set("Retry-After", retryAfter)

	app.respondWithError(w, r, http.StatusTooManyRequests, nil, "too many requests")
}
