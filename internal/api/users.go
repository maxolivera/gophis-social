package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/maxolivera/gophis-social-network/internal/database"
	"golang.org/x/crypto/bcrypt"
)

func (app *Application) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	// types for JSON's input and output
	type input struct {
		Username  string `json:"username"`
		Email     string `json:"email"`
		Password  string `json:"password"`
		FirstName string `json:"first_name,omitempty"`
		LastName  string `json:"last_name,omitempty"`
	}
	type output struct {
		Username string `json:"username"`
		Email    string `json:"email"`
	}
	in := input{}

	// decode input
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&in); err != nil {
		log.Printf("error during JSON decoding: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// check if some parameter was null
	// NOTE(maolivera): should also check password standards or only leave it on client side?
	if in.Username == "" || in.Email == "" || in.Password == "" {
		log.Printf("username, email, and password are required: %s\n", in)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// create user
	id := uuid.New()
	pgID := pgtype.UUID{
		Bytes: id,
		Valid: true,
	}
	currentTime := time.Now().UTC()
	pgTime := pgtype.Timestamp{Time: currentTime, Valid: true}

	// TODO(maolivera): validate password no more than 72 bytes long
	hashed, err := bcrypt.GenerateFromPassword([]byte(in.Password), 14)
	if err != nil {
		log.Println("error when hashing password: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	userParams := database.CreateUserParams{
		ID:        pgID,
		CreatedAt: pgTime,
		UpdatedAt: pgTime,
		Username:  in.Username,
		Email:     in.Email,
		Password:  hashed,
	}
	if in.FirstName != "" {
		userParams.FirstName = in.FirstName
	}
	if in.LastName != "" {
		userParams.LastName = in.LastName
	}

	// store user
	user, err := app.Database.CreateUser(
		r.Context(),
		userParams,
	)
	if err != nil {
		log.Println("error during user creation:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Marshal response
	out := output{}
	out.Username = user.Username
	out.Email = user.Email

	data, err := json.Marshal(out)
	if err != nil {
		log.Printf("failed to marshal JSON response: %v\n", out)
		log.Printf("error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Send response
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
