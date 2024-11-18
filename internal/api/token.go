package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/maxolivera/gophis-social-network/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type CreateTokenPayload struct {
	Email    string
	Password string
}

type TokenResponse struct {
	Token string `json:"token"`
}

// Create Token godoc
//
//	@Summary		Creates a Token
//	@Description	Creates a Token for the user
//	@Tags			authorization
//	@Accept			json
//	@Produce		json
//	@Param			Payload	body		CreateTokenPayload	true	"User credentials"
//	@Success		201		{object}	TokenResponse		"Token"
//	@Failure		500		{object}	error				"Something went wrong on the server"
//	@Failure		409		{object}	error				"Either email or username already taken"
//	@Failure		400		{object}	error				"Some parameter was either not provided or invalid."
//	@Router			/token [post]
func (app *Application) handlerCreateToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	in := CreateTokenPayload{}
	if err := readJSON(w, r, &in); err != nil {
		err := fmt.Errorf("error reading JSON when creating a user: %v", err)
		app.respondWithError(w, r, http.StatusInternalServerError, err, "")
		return
	}
	{ // Validate input
		// Empty input
		if in.Email == "" || in.Password == "" {
			err := fmt.Errorf("username and password are required: %s\n", in)
			app.respondWithError(w, r, http.StatusBadRequest, err, "something is missing")
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
	dbUser, err := app.Database.GetUserByEmail(ctx, in.Email)
	if err != nil {
		switch err {
		case pgx.ErrNoRows:
			// NOTE(maolivera): Returning 404 is insecure
			app.unauthorizedBasicErrorResponse(w, r, err)
		default:
			app.respondWithError(w, r, http.StatusInternalServerError, err, "")
		}
		return
	}

	// compare password
	err = bcrypt.CompareHashAndPassword(dbUser.Password, []byte(in.Password))
	if err != nil {
		err := fmt.Errorf("error when hashing password: %v", err)
		app.unauthorizedBasicErrorResponse(w, r, err)
		return
	}

	user := models.DBUserToUser(dbUser)

	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(app.Config.Authentication.Token.ExpirationTime).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
		"iss": app.Config.Authentication.Token.Issuer,
		"aud": app.Config.Authentication.Token.Issuer,
	}

	token, err := app.Authenticator.GenerateToken(claims)
	if err != nil {
		app.respondWithError(w, r, http.StatusInternalServerError, err, "")
		return
	}

	out := &TokenResponse{token}

	// Send response
	app.respondWithJSON(w, r, http.StatusCreated, out)
}
