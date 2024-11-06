package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/maxolivera/gophis-social-network/internal/database"
)

func (app *Application) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	// types for JSON's input and output
	type parameters struct {
		Username string `json:"username"`
	}
	param := parameters{}

	// decode input
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&param); err != nil {
		log.Printf("error during JSON decoding: %v", err)
		w.WriteHeader(500)
		return
	}

	// check if name was null
	if param.Username == "" {
		log.Printf("user did not provide username", param)
		w.WriteHeader(500)
		return
	}

	currentTime := time.Now().UTC()

	// create user
	u := uuid.New()

	user, err := app.Database.CreateUser(
		r.Context(),
		database.CreateUserParams{
			/*
	// TODO(maolivera): look how to use UUID in sqlc + pgx
	ID			pgtype.UUID
	CreatedAt    pgtype.Timestamp
	UpdatedAt    pgtype.Timestamp
	Username     string
	Email        string
	PasswordHash string
			*/
		},
	)
	if err != nil {
		log.Println("error during user creation:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// TODO(maolivera): return user
	w.WriteHeader(http.StatusOK)
}
