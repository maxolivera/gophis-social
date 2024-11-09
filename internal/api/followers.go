package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/maxolivera/gophis-social-network/internal/database"
)

// TODO(maolivera): Better responses if already following; unfollow someone not already following; etc.

func (app *Application) handlerFollowUser(w http.ResponseWriter, r *http.Request) {
	// TODO(maolivera): Change to auth
	type Follower struct {
		Username string `json:"username"`
	}
	in := &Follower{}
	ctx := r.Context()

	if err := readJSON(w, r, in); err != nil {
		err = fmt.Errorf("error during follower username unmarshal: %v", err)
		respondWithError(w, r, http.StatusInternalServerError, err, "")
		return
	}
	follower, err := app.Database.GetUserByUsername(ctx, in.Username)
	if err != nil {
		err = fmt.Errorf("error during follower retrieve: %v", err)
		respondWithError(w, r, http.StatusInternalServerError, err, "")
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
		respondWithError(w, r, http.StatusInternalServerError, err, "")
		return
	}
}

func (app *Application) handlerUnfollowUser(w http.ResponseWriter, r *http.Request) {
	// TODO(maolivera): Change to auth
	type Follower struct {
		Username string `json:"username"`
	}
	in := &Follower{}
	ctx := r.Context()

	if err := readJSON(w, r, in); err != nil {
		err = fmt.Errorf("error during follower username unmarshal: %v", err)
		respondWithError(w, r, http.StatusInternalServerError, err, "")
		return
	}
	follower, err := app.Database.GetUserByUsername(ctx, in.Username)
	if err != nil {
		err = fmt.Errorf("error during follower retrieve: %v", err)
		respondWithError(w, r, http.StatusInternalServerError, err, "")
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
		respondWithError(w, r, http.StatusInternalServerError, err, "")
		return
	}
}
