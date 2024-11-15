package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/maxolivera/gophis-social-network/internal/database"
	"github.com/maxolivera/gophis-social-network/internal/models"
	// "github.com/maxolivera/gophis-social-network/internal/models"
)

// Feed godoc
//
//	@Summary		Fetches the feed for current user
//	@Description	Fetches the feed for user_id (now, passed as payload, in future by authentication
//	@tags			feed
//	@Accept			json
//	@Produce		json
//	@Param			user_id	header		uuid	true	"User ID"
//	@Param			sort	header		bool	false	"Sort, true if descending order"
//	@Param			limit	header		int32	false	"Limit, 20 by default"
//	@Param			offset	header		int32	false	"Offset, 0 by default"
//	@Success		200		{object}	[]models.Feed
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/v1/ [get]
func (app *Application) handlerFeed(w http.ResponseWriter, r *http.Request) {
	// TODO(maolivera): Change to auth
	ctx := r.Context()

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	type input struct {
		UserID uuid.UUID `json:"user_id"`
		Sort   bool      `json:"sort"`
		Limit  int32     `json:"limit"`
		Offset int32     `json:"offset"`
	}
	in := input{}
	if err := readJSON(w, r, &in); err != nil {
		err := fmt.Errorf("error when reading feed parameters payload: %v", err)
		app.respondWithError(w, r, http.StatusInternalServerError, err, "")
		return
	}

	pgID := pgtype.UUID{Bytes: in.UserID, Valid: true}

	params := database.GetUserFeedParams{
		UserID: pgID,
		Sort:   in.Sort,
		Limit:  in.Limit,
		Offset: in.Offset,
	}

	dbFeed, err := app.Database.GetUserFeed(ctx, params)
	if err != nil {
		// TODO(maolivera): Better error messages and HTTP responses
		err := fmt.Errorf("error retrieving feed for user_id %v, err: %v", in.UserID, err)
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
