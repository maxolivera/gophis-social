package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/maxolivera/gophis-social-network/internal/database"
	"github.com/maxolivera/gophis-social-network/internal/models"
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

	in := FeedPayload{}
	if err := readJSON(w, r, &in); err != nil {
		err := fmt.Errorf("error when reading feed parameters payload: %v", err)
		app.respondWithError(w, r, http.StatusInternalServerError, err, "")
		return
	}

	user := getLoggedUser(r)
	pgID := pgtype.UUID{Bytes: user.ID, Valid: true}

	params := database.GetUserFeedParams{
		UserID: pgID,
		Sort:   in.Sort,
		Limit:  in.Limit,
		Offset: in.Offset,
	}

	dbFeed, err := app.Database.GetUserFeed(ctx, params)
	if err != nil {
		// TODO(maolivera): Better error messages and HTTP responses
		err := fmt.Errorf("error retrieving feed for user %v, err: %v", user.Username, err)
		app.respondWithError(w, r, http.StatusInternalServerError, err, "")
		return
	}

	feed, err := models.DBFeedsToFeeds(dbFeed)
	if err != nil {
		err = fmt.Errorf("error during parsing: %v", err)
		app.respondWithError(w, r, http.StatusInternalServerError, err, "")
		return
	}

	app.respondWithJSON(w, r, http.StatusOK, feed)
}
