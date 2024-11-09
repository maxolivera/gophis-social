package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/maxolivera/gophis-social-network/internal/database"
	"github.com/maxolivera/gophis-social-network/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func (app *Application) middlewareUserContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		username := r.PathValue("username")
		if username == "" {
			err := fmt.Errorf("username not provided")
			// TODO(maolivera): maybe another message?
			respondWithError(w, r, http.StatusBadRequest, err, err.Error())
			return
		}

		dbUser, err := app.Database.GetUserByUsername(
			ctx,
			username,
		)

		if err != nil {
			switch err {
			case pgx.ErrNoRows:
				err := fmt.Errorf("username not found: %v", err)
				respondWithError(w, r, http.StatusNotFound, err, "user not found")
			default:
				respondWithError(w, r, http.StatusInternalServerError, err, "")
			}
			return
		}
		user := models.DBUserToUser(dbUser)

		ctx = context.WithValue(ctx, "user", &user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getUserFromCtx(r *http.Request) *models.User {
	return r.Context().Value("user").(*models.User)
}

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
		err := fmt.Errorf("error when hashing password: %v", err)
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

func (app *Application) handlerGetUser(w http.ResponseWriter, r *http.Request) {
	type output struct {
		ID        uuid.UUID `json:"id"`
		Username  string    `json:"username"`
		Email     string    `json:"email"`
		FirstName string    `json:"first_name"`
		LastName  string    `json:"last_name"`
		CreatedAt time.Time `json:"created_at"`
	}
	user := getUserFromCtx(r)
	out := output{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}
	if user.FirstName != "" {
		out.FirstName = user.FirstName
	}
	if user.LastName != "" {
		out.LastName = user.LastName
	}
	respondWithJSON(w, http.StatusOK, &out)
}

func (app *Application) handlerSoftDeleteUser(w http.ResponseWriter, r *http.Request) {
	user := getUserFromCtx(r)

	pgID := pgtype.UUID{
		Bytes: user.ID,
		Valid: true,
	}

	deleted, err := app.Database.SoftDeleteUserByID(r.Context(), pgID)
	if err != nil {
		switch err {
		case pgx.ErrNoRows:
			err := fmt.Errorf("user was not deleted because was not found, user_id: %v", user.ID)
			respondWithError(w, r, http.StatusNotFound, err, "user not found")
		default:
			err := fmt.Errorf("user could not deleted: %v", err)
			respondWithError(w, r, http.StatusInternalServerError, err, "user could not be deleted")
		}
		return
	}
	if !deleted {
		err := fmt.Errorf("user was not deleted, post_id: %v", user.ID)
		respondWithError(w, r, http.StatusNotFound, err, "user not found")
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}

func (app *Application) handlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	type input struct {
		Email     string `json:"email,omitempty"`
		Username  string `json:"username,omitempty"`
		Password  string `json:"password,omitempty"`
		FirstName string `json:"first_name,omitempty"`
		LastName  string `json:"last_name,omitempty"`
	}
	user := getUserFromCtx(r)

	in := input{}
	if err := readJSON(w, r, &in); err != nil {
		err = fmt.Errorf("error reading input parameters: %v", err)
		respondWithError(w, r, http.StatusInternalServerError, err, "")
	}

	currentTime := time.Now().UTC()
	pgTime := pgtype.Timestamp{Time: currentTime, Valid: true}
	pgEmail := pgtype.Text{String: in.Email, Valid: len(in.Email) > 0}
	pgUsername := pgtype.Text{String: in.Username, Valid: len(in.Username) > 0}
	pgFirstName := pgtype.Text{String: in.FirstName, Valid: len(in.FirstName) > 0}
	pgLastName := pgtype.Text{String: in.LastName, Valid: len(in.LastName) > 0}
	var pgPassword []byte
	if in.Password != "" {
		// TODO(maolivera): validate password no more than 72 bytes long
		hashed, err := bcrypt.GenerateFromPassword([]byte(in.Password), 14)
		if err != nil {
			err := fmt.Errorf("error when hashing password: %v", err)
			respondWithError(w, r, http.StatusInternalServerError, err, "")
			return
		}
		pgPassword = hashed
	} else {
		pgPassword = nil
	}
	pgID := pgtype.UUID{Bytes: user.ID, Valid: true}

	params := database.UpdateUserParams{
		UpdatedAt: pgTime,
		ID:        pgID,
		Username:  pgUsername,
		Email:     pgEmail,
		FirstName: pgFirstName,
		LastName:  pgLastName,
		Password:  pgPassword,
	}

	newDBUser, err := app.Database.UpdateUser(r.Context(), params)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			switch pgErr.ConstraintName {
			case "users_username_key":
				msg := "username not available"
				err = fmt.Errorf("%s: %v", msg, err)
				respondWithError(w, r, http.StatusConflict, err, msg)
			case "users_email_key":
				msg := "email not available"
				err = fmt.Errorf("%s: %v", msg, err)
				respondWithError(w, r, http.StatusConflict, err, msg)
			default:
				msg := "something went wrong"
				err = fmt.Errorf("%s: %v", msg, err)
				respondWithError(w, r, http.StatusConflict, err, msg)
			}
			return
		}
		switch err {
		case pgx.ErrNoRows:
			respondWithError(w, r, http.StatusNotFound, err, "user not found")
		default:
			err := fmt.Errorf("post could not updated: %v", err)
			respondWithError(w, r, http.StatusInternalServerError, err, "")
		}
		return
	}
	newUser := models.DBUserToUser(newDBUser)

	respondWithJSON(w, http.StatusOK, newUser)
}
