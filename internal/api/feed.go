package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/maxolivera/gophis-social-network/internal/database"
	// "github.com/maxolivera/gophis-social-network/internal/models"
)

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
		respondWithError(w, r, http.StatusInternalServerError, err, "")
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
		respondWithError(w, r, http.StatusInternalServerError, err, "")
		return
	}
	respondWithJSON(w, http.StatusOK, dbFeed)

	/*
	feed, err := models.DBFeedsToFeeds(dbFeed)
	if err != nil {
		err = fmt.Errorf("error during parsing: %v", err)
		respondWithError(w, r, http.StatusInternalServerError, err, "")
		return
	}

	respondWithJSON(w, http.StatusOK, feed)
	*/
}
