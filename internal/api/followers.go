package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/maxolivera/gophis-social-network/internal/database"
)

// TODO(maolivera): Better responses if already following; unfollow someone not already following; etc.

// Follow godoc
//
//	@Summary		Follows an User
//	@Description	Logged user (identified by username in header, in future with auth) will start following user at /{username}. This is an idempotent endpoint, which means that it will always produce the same result, or in other words, if some user tries to follow someone who is already following it, nothing will happen
//	@tags			users
//	@Accept			json
//	@Produce		json
//	@Param			username	path	string	true	"Username of user to follow"
//	@Param			username	header	string	true	"Follower username"
//	@Success		204			"Follower will follow username"
//	@Failure		500			{object}	error
//	@Failure		404			{object}	error	"User at /{username} was not found"
//	@Failure		400			{object}	error	"Unlikely. Username at /{username} was not provided"
//	@Security		ApiKeyAuth
//	@Router			/v1/{username}/follow [put]
func (app *Application) handlerFollowUser(w http.ResponseWriter, r *http.Request) {
	// TODO(maolivera): Change to auth
	type Follower struct {
		Username string `json:"username"`
	}
	in := &Follower{}
	ctx := r.Context()

	if err := readJSON(w, r, in); err != nil {
		err = fmt.Errorf("error during follower username unmarshal: %v", err)
		app.respondWithError(w, r, http.StatusInternalServerError, err, "")
		return
	}
	follower, err := app.Database.GetUserByUsername(ctx, in.Username)
	if err != nil {
		err = fmt.Errorf("error during follower retrieve: %v", err)
		app.respondWithError(w, r, http.StatusInternalServerError, err, "")
		return
	}
	// ==

	user := getUserFromCtx(r)

	currentTime := time.Now().UTC()
	pgTime := pgtype.Timestamp{Time: currentTime, Valid: true}
	pgUserID := pgtype.UUID{Bytes: user.ID, Valid: true}
	pgFollowerID := follower.ID

	params := database.FollowByIDParams{
		CreatedAt:  pgTime,
		UserID:     pgUserID,
		FollowerID: pgFollowerID,
	}

	if err := app.Database.FollowByID(ctx, params); err != nil {
		err = fmt.Errorf("error during following user: %v", err)
		app.respondWithError(w, r, http.StatusInternalServerError, err, "")
		return
	}

	app.respondWithJSON(w, r, http.StatusNoContent, nil)
}

// Unfollow godoc
//
//	@Summary		Unfollows an User
//	@Description	Logged user (identified by username in header, in future with auth) will start following user at /{username}. This is an idempotent endpoint, which means that it will always produce the same result, or in other words, if some user tries to follow someone who is already following it, nothing will happen
//	@tags			users
//	@Accept			json
//	@Produce		json
//	@Param			username	path	string	true	"Username of user to unfollow"
//	@Param			username	header	string	true	"Follower username"
//	@Success		204			"Follower will unfollow username"
//	@Failure		500			{object}	error
//	@Failure		404			{object}	error	"User at /{username} was not found"
//	@Failure		400			{object}	error	"Unlikely. Username at /{username} was not provided"
//	@Security		ApiKeyAuth
//	@Router			/v1/{username}/unfollow [put]
func (app *Application) handlerUnfollowUser(w http.ResponseWriter, r *http.Request) {
	// TODO(maolivera): Change to auth
	type Follower struct {
		Username string `json:"username"`
	}
	in := &Follower{}
	ctx := r.Context()

	if err := readJSON(w, r, in); err != nil {
		err = fmt.Errorf("error during follower username unmarshal: %v", err)
		app.respondWithError(w, r, http.StatusInternalServerError, err, "")
		return
	}
	follower, err := app.Database.GetUserByUsername(ctx, in.Username)
	if err != nil {
		err = fmt.Errorf("error during follower retrieve: %v", err)
		app.respondWithError(w, r, http.StatusInternalServerError, err, "")
		return
	}
	// ==

	user := getUserFromCtx(r)

	pgUserID := pgtype.UUID{Bytes: user.ID, Valid: true}
	pgFollowerID := follower.ID

	params := database.UnfollowByIDParams{
		UserID:     pgUserID,
		FollowerID: pgFollowerID,
	}

	if err := app.Database.UnfollowByID(ctx, params); err != nil {
		err = fmt.Errorf("error during following user: %v", err)
		app.respondWithError(w, r, http.StatusInternalServerError, err, "")
		return
	}
}
