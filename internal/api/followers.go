package api

import (
	"fmt"
	"net/http"
)

// Follow godoc
//
//	@Summary		Follows an User
//	@Description	Logged user will start following user at /{username}. This is an idempotent endpoint, which means that it will always produce the same result, or in other words, if some user tries to follow someone who is already following it, nothing will happen
//	@tags			users
//	@Accept			json
//	@Produce		json
//	@Param			username	path	string	true	"User to follow"
//	@Success		204			"Follower will follow username"
//	@Failure		500			{object}	error
//	@Failure		404			{object}	error	"User at /{username} was not found"
//	@Security		ApiKeyAuth
//	@Router			/v1/{username}/follow [put]
func (app *Application) handlerFollowUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	routeUser := getRouteUser(r)
	loggedUser := getLoggedUser(r)

	if err := app.Storage.Followers.Follow(ctx, routeUser.ID, loggedUser.ID); err != nil {
		err = fmt.Errorf("error during following user: %v", err)
		app.respondWithError(w, r, http.StatusInternalServerError, err, "")
		return
	}

	app.respondWithJSON(w, r, http.StatusNoContent, nil)
}

// Unfollow godoc
//
//	@Summary		Unfollows an User
//	@Description	Logged user will start following user at /{username}. This is an idempotent endpoint, which means that it will always produce the same result, or in other words, if some user tries to follow someone who is already following it, nothing will happen
//	@tags			users
//	@Accept			json
//	@Produce		json
//	@Param			username	path	string	true	"User to unfollow"
//	@Success		204			"Follower will unfollow username"
//	@Failure		500			{object}	error
//	@Failure		404			{object}	error	"User at /{username} was not found"
//	@Security		ApiKeyAuth
//	@Router			/v1/{username}/unfollow [put]
func (app *Application) handlerUnfollowUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	routeUser := getRouteUser(r)
	loggedUser := getLoggedUser(r)

	if err := app.Storage.Followers.Follow(ctx, routeUser.ID, loggedUser.ID); err != nil {
		err = fmt.Errorf("error during following user: %v", err)
		app.respondWithError(w, r, http.StatusInternalServerError, err, "")
		return
	}
}
