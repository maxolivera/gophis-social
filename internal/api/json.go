package api

import (
	"encoding/json"
	"log"
	"net/http"
)

const MAX_BYTES = 1_048_578 // 1 MB

func respondWithJSON(w http.ResponseWriter, code int, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal JSON response: %v", payload)
		log.Printf("Error: %v", err)

		// NOTE(maolivera): Hardcoding to avoid adding http.Request parameter. Maybe adding and re-use function?

		w.Write([]byte("{\"error\": \"the server encountered a problem\" }"))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(data)
	return
}

// Will sent a response to the client with code `code` and message `message`, and will log the error `err`
func respondWithError(w http.ResponseWriter, r *http.Request, code int, err error, message string) {
	type response struct {
		Error string `json:"error"`
	}

	res := response{}

	// default messages for specific codes
	if message == "" {
		switch code {
		case http.StatusInternalServerError:
			res.Error = "the server encountered a problem"
		default:
			log.Printf("the HTTP code %d do not support default message, an empty message will be sent", code)
			res.Error = ""
		}
	} else {
		res.Error = message
	}

	log.Printf("sending: %d on method %s, path: %s error: %s\n", code, r.Method, r.URL.Path, err)

	respondWithJSON(w, code, res)
}

func readJSON(w http.ResponseWriter, r *http.Request, data any) error {
	r.Body = http.MaxBytesReader(w, r.Body, MAX_BYTES)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	return decoder.Decode(data)
}
