package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/maxolivera/gophis-social-network/internal/storage"
)

type FeedPayload struct {
	Sort   bool  `json:"sort"`
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

// Feed godoc
//
//	@Summary		Fetches the feed for current user
//	@Description	Fetches the feed for user_id (now, passed as payload, in future by authentication
//	@tags			feed
//	@Accept			json
//	@Produce		json
//	@Param			Payload	body		FeedPayload	true	"Feed payload"
//	@Success		200		{object}	[]models.Feed
//	@Success		401		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/v1/feed [get]
func (app *Application) handlerFeed(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
	defer cancel()
	user := getLoggedUser(r)

	in := FeedPayload{}
	if err := readJSON(w, r, &in); err != nil {
		err := fmt.Errorf("error when reading feed parameters payload: %v", err)
		app.respondWithError(w, r, http.StatusInternalServerError, err, "")
		return
	}

	feed, err := app.Storage.Posts.GetFeed(ctx, user, in.Sort, in.Limit, in.Offset)
	if err != nil {
		if err == storage.ErrNoRows {
			err = fmt.Errorf("no posts for user %s, with limit %d and offset %d", user.Username, in.Limit, in.Offset)
			app.respondWithError(w, r, http.StatusNotFound, err, "no posts for feed")
		} else {
			err := fmt.Errorf("error retrieving feed for user %v, err: %v", user.Username, err)
			app.respondWithError(w, r, http.StatusInternalServerError, err, "")
		}
		return
	}

	app.respondWithJSON(w, r, http.StatusOK, feed)
}
