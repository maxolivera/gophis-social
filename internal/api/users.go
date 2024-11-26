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
	"github.com/maxolivera/gophis-social-network/internal/storage"
	"github.com/maxolivera/gophis-social-network/internal/storage/models"
)

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
			app.respondWithError(w, r, http.StatusBadRequest, err, "username, email and password are required")
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

	// Create user
	id := uuid.New()
	currentTime := time.Now().UTC()
	user := &models.UserWithPassword{
		User: models.User{
			ID:        id,
			CreatedAt: currentTime,
			UpdatedAt: currentTime,
			Email:     in.Email,
			Username:  in.Username,
		},
		Password: in.Password,
	}

	// Create Invitation Token
	token := make([]byte, 32)
	_, err := rand.Read(token)
	if err != nil {
		err = fmt.Errorf("error creating token: %v", err)
		app.respondWithError(w, r, http.StatusInternalServerError, err, "")
		return
	}

	if err = app.Storage.Users.CreateAndInvite(ctx, user, token, app.Config.ExpirationTime); err != nil {
		switch err {
		case storage.ErrUsernameUnavailable:
			err = fmt.Errorf("%v, username: %s", err, user.User.Username)
			app.respondWithError(w, r, http.StatusConflict, err, "username is not available")
		case storage.ErrEmailUnavailable:
			err = fmt.Errorf("%v, email: %s", err, user.User.Email)
			app.respondWithError(w, r, http.StatusConflict, err, "email is not available")
		default:
			err = fmt.Errorf("error during user creation: %v", err)
			app.respondWithError(w, r, http.StatusInternalServerError, err, "")
		}
		return
	}

	// 4. Send token
	// TODO(maolivera): Send email instead of returning the user with the token
	// TODO(maolivera): Check if this is the correct way of encoding and decoding token for URLs
	encodedToken := base64.URLEncoding.EncodeToString(token)
	out := UserWithToken{
		User:            user.User,
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

	// Decode token
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

	user, err := app.Storage.Users.Activate(ctx, token)

	if err != nil {
		switch err {
		case storage.ErrNoToken:
			err = fmt.Errorf("token not found or expired: %x", token)
			app.respondWithError(w, r, http.StatusNotFound, err, "token not found or expired")
		case storage.ErrNoUser:
			// NOTE(maolivera): I suppose this could happen if the user tries to activate the user AFTER deleting the account
			app.respondWithError(w, r, http.StatusNotFound, err, "user not found")
		default:
			err = fmt.Errorf("error during user activation: %v", err)
			app.respondWithError(w, r, http.StatusInternalServerError, err, "token not found")
		}
		return
	}
	app.Logger.Infow("user activated", "username", user.Username, "id", fmt.Sprintf("%x", user.ID))

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
//	@Security		ApiKeyAuth
func (app *Application) handlerGetUser(w http.ResponseWriter, r *http.Request) {
	user := getRouteUser(r)
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
//	@Security		ApiKeyAuth
func (app *Application) handlerSoftDeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getRouteUser(r)

	if err := app.Storage.Users.SoftDelete(ctx, user.ID); err != nil {
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

	app.respondWithJSON(w, r, http.StatusNoContent, nil)
}

// Hard Delete User godoc
//
//	@Summary		Hard Deletes a User
//	@Description	The user will be deleted. Currently not used.
//	@Tags			admin, users
//	@Produce		json
//	@Param			username	path	string	true	"Username"
//	@Success		204			"The user was deleted"
//	@Failure		400			{object}	error	"Some parameter was either not provided or invalid."
//	@Failure		404			{object}	error	"User not found"
//	@Failure		500			{object}	error	"Something went wrong on the server"
//	@Router			/users/{username}/hard [delete]
func (app *Application) handlerHardDeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getRouteUser(r)

	if err := app.Storage.Users.HardDelete(ctx, user.ID); err != nil {
		switch err {
		case pgx.ErrNoRows:
			err := fmt.Errorf("User not deleted. Not found, user id: %v error: %v", user.ID, err)
			app.respondWithError(w, r, http.StatusNotFound, err, "user not found")
		default:
			err := fmt.Errorf("User could not deleted: %v", err)
			app.respondWithError(w, r, http.StatusInternalServerError, err, "user could not be deleted")
		}
		return
	}

	app.respondWithJSON(w, r, http.StatusNoContent, nil)
}

type UpdateUserPayload struct {
	Email     string `json:"email,omitempty"`
	Username  string `json:"username,omitempty"`
	Password  string `json:"password,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

// Update User godoc
//
//	@Summary		Updates an User
//	@Description	The user will be updated.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			username	path		string				true	"Username of user to be modified"
//	@Param			Payload		body		UpdateUserPayload	false	"New parameters for user"
//	@Success		200			{object}	models.User			"New User"
//	@Failure		404			{object}	error				"User not found"
//	@Failure		400			{object}	error				"Some parameter was either not provided or invalid."
//	@Failure		409			{object}	error				"Either email or username already taken"
//	@Failure		500			{object}	error				"Something went wrong on the server"
//	@Router			/users/{username} [patch]
//	@Security		ApiKeyAuth
func (app *Application) handlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getRouteUser(r)

	in := UpdateUserPayload{}
	if err := readJSON(w, r, &in); err != nil {
		err = fmt.Errorf("error reading input parameters: %v", err)
		app.respondWithError(w, r, http.StatusInternalServerError, err, "")
	}

	// TODO(maolivera): Split update password to its own endpoint
	// TODO(maolivera): Better validation?
	{ // Validate input
		// Username
		if len(in.Username) > 100 {
			err := errors.New("username is too long")
			app.respondWithError(w, r, http.StatusBadRequest, err, "username is too long")
			return
		}
		if len(in.FirstName) > 100 {
			err := errors.New("first name is too long")
			app.respondWithError(w, r, http.StatusBadRequest, err, "first name is too long")
			return
		}
		if len(in.LastName) > 100 {
			err := errors.New("last name is too long")
			app.respondWithError(w, r, http.StatusBadRequest, err, "last name is too long")
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

	newUser := &models.UserWithPassword{
		User: models.User{
			Email:     in.Email,
			Username:  in.Username,
			FirstName: in.FirstName,
			LastName:  in.LastName,
		},
		Password: in.Password,
	}

	changedUser, err := app.Storage.Users.Update(ctx, newUser)
	if err != nil {
		switch err {
		case storage.ErrUsernameUnavailable:
			err = fmt.Errorf("%v, username: %s", err, newUser.User.Username)
			app.respondWithError(w, r, http.StatusConflict, err, "username is not available")
		case storage.ErrEmailUnavailable:
			err = fmt.Errorf("%v, email: %s", err, newUser.User.Email)
			app.respondWithError(w, r, http.StatusConflict, err, "email is not available")
		default:
			err = fmt.Errorf("error during user creation: %v", err)
			app.respondWithError(w, r, http.StatusInternalServerError, err, "")
		}
		return
	}

	if app.Config.Cache.Enabled {
		app.Cache.Users.Delete(r.Context(), user.Username)
	}

	app.respondWithJSON(w, r, http.StatusOK, changedUser)
}
