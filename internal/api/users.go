package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
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
		Username string    `json:"username"`
		ID       uuid.UUID `json:"id"`
		Email    string    `json:"email"`
	}
	in := input{}

	if err := readJSON(w, r, &in); err != nil {
		err := fmt.Errorf("error reading JSON when creating a user: %v", err)
		respondWithError(w, r, http.StatusInternalServerError, err, "")
		return
	}

	// check if some parameter was null
	// NOTE(maolivera): should also check password standards or only leave it on client side?
	if in.Username == "" || in.Email == "" || in.Password == "" {
		err := fmt.Errorf("username, email, and password are required: %s\n", in)
		respondWithError(w, r, http.StatusBadRequest, err, "something is missing")
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
		err := fmt.Errorf("error when hashing password: ", err)
		respondWithError(w, r, http.StatusInternalServerError, err, "")
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
		pgFirstName := pgtype.Text{
			String: in.FirstName,
			Valid:  true,
		}
		userParams.FirstName = pgFirstName
	}
	if in.LastName != "" {
		pgLastName := pgtype.Text{
			String: in.LastName,
			Valid:  true,
		}
		userParams.LastName = pgLastName
	}

	// store user
	user, err := app.Database.CreateUser(
		r.Context(),
		userParams,
	)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			switch pgErr.ConstraintName {
			case "users_username_key":
				msg := "username not available"
				err = fmt.Errorf("%s: %v", msg, err)
				respondWithError(w, r, http.StatusConflict, err, msg)
				return
			case "users_email_key":
				msg := "email not available"
				err = fmt.Errorf("%s: %v", msg, err)
				respondWithError(w, r, http.StatusConflict, err, msg)
				return
			default:
				err = fmt.Errorf("error during user creation: %v", err)
			}
		} else {
			err = fmt.Errorf("error during user creation: %v", err)
		}
		respondWithError(w, r, http.StatusInternalServerError, err, "")
		return
	}

	// Marshal response
	out := output{}
	out.Username = user.Username
	out.Email = user.Email
	out.ID = user.ID.Bytes

	// Send response
	respondWithJSON(w, http.StatusOK, out)
}
