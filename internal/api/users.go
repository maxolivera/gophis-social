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

func getUserFromCtx(r *http.Request) models.User {
	return r.Context().Value("user").(models.User)
}

// Create User godoc
//
//	@Summary		Creates a User
//	@Description	Created a User
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			Username	header		string	true	"Username"
//	@Param			Email		header		string	true	"Email"
//	@Param			Password	header		string	true	"Password"
//	@Param			FirstName	header		string	false	"First Name"
//	@Param			LastName	header		string	false	"Last Name"
//	@Success		200			{object}	models.User
//	@Failure		500			{object}	error	"Something went wrong on the server"
//	@Failure		409			{object}	error	"Either email or username already taken"
//	@Failure		400			{object}	error	"Some parameter was either not provided or invalid."
//	@Router			/users [post]
func (app *Application) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	// types for JSON's input and output
	type input struct {
		Username  string `json:"username"`
		Email     string `json:"email"`
		Password  string `json:"password"`
		FirstName string `json:"first_name,omitempty"`
		LastName  string `json:"last_name,omitempty"`
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
	dbUser, err := app.Database.CreateUser(r.Context(), userParams)
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

	user := models.DBUserToUser(dbUser)

	// Send response
	respondWithJSON(w, http.StatusOK, user)
}

// Get User godoc
//
//	@Summary		Fetch a User
//	@Description	Fetch a User
//	@Tags			users
//	@Produce		json
//	@Param			username	path		string	true	"Username"
//	@Success		200			{object}	models.User
//	@Failure		404			{object}	error	"User not found"
//	@Failure		400			{object}	error	"Some parameter was either not provided or invalid."
//	@Failure		500			{object}	error	"Something went wrong on the server"
//	@Router			/users/{username} [get]
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
	respondWithJSON(w, http.StatusOK, user)
}

// Soft Delete User godoc
//
//	@Summary		Soft Deletes a User
//	@Description	The user will be marked as "deleted" on the database, it will not appear nor it can be accessed, but it will not be deleted from the database
//	@Tags			users
//	@Produce		json
//	@Param			username	path	string	true	"Username"
//	@Success		204			"The user was deleted"
//	@Failure		400			{object}	error	"Some parameter was either not provided or invalid."
//	@Failure		404			{object}	error	"User not found"
//	@Failure		500			{object}	error	"Something went wrong on the server"
//	@Router			/users/{username} [delete]
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

/*
// Hard Delete User godoc
//
//	@Summary		Hard Deletes a User
//	@Description	The user will be deleted. Currently not used.
//	@Tags			admin, users
//	@Produce		json
//	@Param			username path string true "Username"
//	@Success		204	"The user was deleted"
//	@Failure		400	{object}	error "Some parameter was either not provided or invalid."
//	@Failure		404	{object}	error "User not found"
//	@Failure		500	{object}	error "Something went wrong on the server"
//	@Router			/users/{username} [delete]
// TODO(maolivera): Implement Hard Delete user
*/

// Update User godoc
//
//	@Summary		Updates an User
//	@Description	The user will be updated. In future will require auth.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			username	path		string		true	"Username of user to be modified"
//	@Param			Email		header		string		false	"New email"
//	@Param			Username	header		string		false	"New username"
//	@Param			Password	header		string		false	"New password"
//	@Param			FirstName	header		string		false	"New first name"
//	@Param			LastName	header		string		false	"New last name"
//	@Success		200			{object}	models.User	"New User"
//	@Failure		404			{object}	error		"User not found"
//	@Failure		400			{object}	error		"Some parameter was either not provided or invalid."
//	@Failure		409			{object}	error		"Either email or username already taken"
//	@Failure		500			{object}	error		"Something went wrong on the server"
//	@Router			/users/{username} [patch]
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

	// TODO(maolivera): Validate fields

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
