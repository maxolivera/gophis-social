package api

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"net/url"
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
			app.respondWithError(w, r, http.StatusBadRequest, err, err.Error())
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
				app.respondWithError(w, r, http.StatusNotFound, err, "user not found")
			default:
				app.respondWithError(w, r, http.StatusInternalServerError, err, "")
			}
			return
		}
		user := models.DBUserToUser(dbUser)

		ctx = context.WithValue(ctx, "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getUserFromCtx(r *http.Request) models.User {
	return r.Context().Value("user").(models.User)
}

type CreateUserPayload struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserWithToken struct {
	User            models.User `json:"user"`
	ActivationToken string      `json:"activation_token"`
}

// Create User godoc
//
//	@Summary		Creates a User
//	@Description	Creates a User
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			Payload	body		CreateUserPayload	true	"User credentials"
//	@Success		201		{object}	UserWithToken
//	@Failure		500		{object}	error	"Something went wrong on the server"
//	@Failure		409		{object}	error	"Either email or username already taken"
//	@Failure		400		{object}	error	"Some parameter was either not provided or invalid."
//	@Router			/register [post]
func (app *Application) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	in := CreateUserPayload{}
	if err := readJSON(w, r, &in); err != nil {
		err := fmt.Errorf("error reading JSON when creating a user: %v", err)
		app.respondWithError(w, r, http.StatusInternalServerError, err, "")
		return
	}

	{ // Validate input
		// Empty input
		if in.Username == "" || in.Email == "" || in.Password == "" {
			err := fmt.Errorf("username, email, and password are required: %s\n", in)
			app.respondWithError(w, r, http.StatusBadRequest, err, "something is missing")
			return
		}
		// Username
		if len(in.Username) > 100 {
			err := errors.New("username is too long")
			app.respondWithError(w, r, http.StatusBadRequest, err, "username is too long")
			return
		}
		// Email
		if len(in.Email) > 255 {
			err := errors.New("email is too long")
			app.respondWithError(w, r, http.StatusBadRequest, err, "email is too long")
			return
		}
		if _, err := mail.ParseAddress(in.Email); err != nil {
			err := fmt.Errorf("email is invalid: %v", err)
			app.respondWithError(w, r, http.StatusBadRequest, err, "email is invalid")
			return
		}
		// Password
		if len(in.Password) > 72 {
			err := errors.New("password is too long")
			app.respondWithError(w, r, http.StatusBadRequest, err, "password is too long")
			return
		}
		if len(in.Password) < 3 {
			err := errors.New("password is too short")
			app.respondWithError(w, r, http.StatusBadRequest, err, "password is too short")
			return
		}
	}

	// hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		err := fmt.Errorf("error when hashing password: %v", err)
		app.respondWithError(w, r, http.StatusInternalServerError, err, "")
		return
	}

	// Start transaction
	tx, err := app.Pool.Begin(ctx)
	if err != nil {
		err := fmt.Errorf("error starting transaction: %v", err)
		app.respondWithError(w, r, http.StatusInternalServerError, err, "")
		return
	}
	defer tx.Rollback(ctx)
	qtx := app.Database.WithTx(tx)

	// 1. Create user
	id := uuid.New()
	pgID := pgtype.UUID{
		Bytes: id,
		Valid: true,
	}
	currentTime := time.Now().UTC()
	pgTime := pgtype.Timestamp{Time: currentTime, Valid: true}

	userParams := database.CreateUserParams{
		ID:        pgID,
		CreatedAt: pgTime,
		UpdatedAt: pgTime,
		Username:  in.Username,
		Email:     in.Email,
		Password:  hashed,
	}
	// store user
	dbUser, err := qtx.CreateUser(r.Context(), userParams)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			switch pgErr.ConstraintName {
			case "users_username_key":
				msg := "username not available"
				err = fmt.Errorf("%s: %v", msg, err)
				app.respondWithError(w, r, http.StatusConflict, err, msg)
				return
			case "users_email_key":
				msg := "email not available"
				err = fmt.Errorf("%s: %v", msg, err)
				app.respondWithError(w, r, http.StatusConflict, err, msg)
				return
			default:
				err = fmt.Errorf("error during user creation: %v", err)
			}
		} else {
			err = fmt.Errorf("error during user creation: %v", err)
		}
		app.respondWithError(w, r, http.StatusInternalServerError, err, "")
		return
	}
	user := models.DBUserToUser(dbUser)

	// 2. Create user invite
	token := make([]byte, 32)
	_, err = rand.Read(token)
	if err != nil {
		err = fmt.Errorf("error creating token: %v", err)
		app.respondWithError(w, r, http.StatusInternalServerError, err, "")
		return
	}
	pgExpires := pgtype.Timestamp{Time: currentTime.Add(app.Config.ExpirationTime), Valid: true}
	invitationParams := database.CreateInvitationParams{
		UserID:    pgID,
		Token:     token,
		ExpiresAt: pgExpires,
	}
	if err := qtx.CreateInvitation(ctx, invitationParams); err != nil {
		err = fmt.Errorf("error storing token: %v", err)
		app.respondWithError(w, r, http.StatusInternalServerError, err, "")
		return
	}

	// 3. Commit transaction
	if err = tx.Commit(ctx); err != nil {
		err = fmt.Errorf("error commiting user creation transaction: %v", err)
		app.respondWithError(w, r, http.StatusInternalServerError, err, "")
		return
	}

	// 4. Send token
	// TODO(maolivera): Send email instead of returning the user with the token
	// TODO(maolivera): Check if this is the correct way of encoding and decoding token for URLs
	encodedToken := base64.URLEncoding.EncodeToString(token)
	out := UserWithToken{
		User:            user,
		ActivationToken: encodedToken,
	}

	// Send response
	app.respondWithJSON(w, r, http.StatusCreated, out)
}

// Activate User godoc
//
//	@Summary		Activates a User
//	@Description	Activates a User
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			token	path	string	true	"Activation token"
//	@Success		204
//	@Failure		500	{object}	error	"Something went wrong on the server"
//	@Failure		400	{object}	error	"Invalid token"
//	@Failure		404	{object}	error	"Token not found"
//	@Router			/activate/{token} [post]
func (app *Application) handlerActivateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// TODO(maolivera): Check if this is the correct way of encoding and decoding token for URLs
	tokenStr := r.PathValue("token")
	if tokenStr == "" {
		err := fmt.Errorf("token required but it was not found, url: %s", r.URL.String())
		app.respondWithError(w, r, http.StatusBadRequest, err, "token not provided/found")
		return
	}
	decodedToken, err := url.QueryUnescape(tokenStr)
	if err != nil {
		err := fmt.Errorf("error unescaping token: %v", err)
		app.respondWithError(w, r, http.StatusBadRequest, err, "invalid token")
		return
	}
	token, err := base64.URLEncoding.DecodeString(decodedToken)
	if err != nil {
		err := fmt.Errorf("error decoding token string: %v", err)
		app.respondWithError(w, r, http.StatusBadRequest, err, "invalid token")
		return
	}

	app.Logger.Debugw("token decoded", "token", fmt.Sprintf("%x", token))

	// 0. Start transaction
	tx, err := app.Pool.Begin(ctx)
	if err != nil {
		err := fmt.Errorf("error starting transaction: %v", err)
		app.respondWithError(w, r, http.StatusInternalServerError, err, "")
		return
	}
	defer tx.Rollback(ctx)
	qtx := app.Database.WithTx(tx)

	// 1. Validate token
	id, err := qtx.GetInvitation(ctx, database.GetInvitationParams{
		Token:     token,
		ExpiresAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
	})
	if err != nil {
		switch err {
		case pgx.ErrNoRows:
			app.respondWithError(w, r, http.StatusNotFound, err, "expired or invalid token")
		default:
			app.respondWithError(w, r, http.StatusInternalServerError, err, "")
		}
		return
	}

	// 2. Activate user
	if err = qtx.ActivateUser(ctx, id); err != nil {
		// NOTE(maolivera): It should exists as the user_id on user_invitations table is "DELETE ON CASCADE", so the user should always exists
		err = fmt.Errorf("error activating user: %v", err)
		app.respondWithError(w, r, http.StatusInternalServerError, err, "")
		return
	}
	app.Logger.Infow("user activated", "user_id", id.Bytes)

	// 3. Delete token from DB
	if err = qtx.DeleteToken(ctx, token); err != nil {
		err = fmt.Errorf("error deleting token: %v", err)
		app.respondWithError(w, r, http.StatusInternalServerError, err, "")
		return
	}

	// 4. Finish transaction
	if err = tx.Commit(ctx); err != nil {
		err = fmt.Errorf("error commiting user activation transaction: %v", err)
		app.respondWithError(w, r, http.StatusInternalServerError, err, "")
		return
	}

	// Send response
	app.respondWithJSON(w, r, http.StatusNoContent, nil)
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
	app.respondWithJSON(w, r, http.StatusOK, user)
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
			app.respondWithError(w, r, http.StatusNotFound, err, "user not found")
		default:
			err := fmt.Errorf("user could not deleted: %v", err)
			app.respondWithError(w, r, http.StatusInternalServerError, err, "user could not be deleted")
		}
		return
	}
	if !deleted {
		err := fmt.Errorf("user was not deleted, post_id: %v", user.ID)
		app.respondWithError(w, r, http.StatusNotFound, err, "user not found")
		return
	}

	app.respondWithJSON(w, r, http.StatusNoContent, nil)
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
		app.respondWithError(w, r, http.StatusInternalServerError, err, "")
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
			app.respondWithError(w, r, http.StatusInternalServerError, err, "")
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
				app.respondWithError(w, r, http.StatusConflict, err, msg)
			case "users_email_key":
				msg := "email not available"
				err = fmt.Errorf("%s: %v", msg, err)
				app.respondWithError(w, r, http.StatusConflict, err, msg)
			default:
				msg := "something went wrong"
				err = fmt.Errorf("%s: %v", msg, err)
				app.respondWithError(w, r, http.StatusConflict, err, msg)
			}
			return
		}
		switch err {
		case pgx.ErrNoRows:
			app.respondWithError(w, r, http.StatusNotFound, err, "user not found")
		default:
			err := fmt.Errorf("post could not updated: %v", err)
			app.respondWithError(w, r, http.StatusInternalServerError, err, "")
		}
		return
	}
	newUser := models.DBUserToUser(newDBUser)

	app.respondWithJSON(w, r, http.StatusOK, newUser)
}
