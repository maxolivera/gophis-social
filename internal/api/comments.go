package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/maxolivera/gophis-social-network/internal/database"
)

func (app *Application) handlerCreateComment(w http.ResponseWriter, r *http.Request) {
	// TODO(maolivera): change this to auth instead of beign passed
	type input struct {
		UserID  uuid.UUID `json:"user_id"`
		PostID  uuid.UUID `json:"post_id"`
		Content string    `json:"content"`
	}
	type output struct {
		ID        uuid.UUID `json:"id"`
		UserID    uuid.UUID `json:"user_id"`
		PostID    uuid.UUID `json:"post_id"`
		Content   string    `json:"content"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}
	in := input{}

	if err := readJSON(w, r, &in); err != nil {
		app.respondWithError(w, r, http.StatusInternalServerError, err, "")
		return
	}
	// validate
	if len(in.Content) > MAX_CONTENT_LENGTH {
		err := fmt.Errorf("content is too long, max is %d vs. current %d", MAX_CONTENT_LENGTH, len(in.Content))
		app.respondWithError(w, r, http.StatusBadRequest, err, "content is too long")
		return
	}
	if in.UserID == uuid.Nil {
		err := fmt.Errorf("user_id not provided, input: %v", in)
		app.respondWithError(w, r, http.StatusBadRequest, err, "not authenticated")
		return
	}
	if in.PostID == uuid.Nil {
		err := fmt.Errorf("post_id not provided, input: %v", in)
		app.respondWithError(w, r, http.StatusBadRequest, err, "invalid post")
		return
	}

	// create post
	id := uuid.New()
	pgID := pgtype.UUID{
		Bytes: id,
		Valid: true,
	}
	pgUserID := pgtype.UUID{
		Bytes: in.UserID,
		Valid: true,
	}
	pgPostId := pgtype.UUID{
		Bytes: in.PostID,
		Valid: true,
	}
	currentTime := time.Now().UTC()
	pgTime := pgtype.Timestamp{Time: currentTime, Valid: true}

	commentsParams := database.CreateCommentInPostParams{
		ID:        pgID,
		PostID:    pgPostId,
		UserID:    pgUserID,
		CreatedAt: pgTime,
		UpdatedAt: pgTime,
		Content:   in.Content,
	}

	comment, err := app.Database.CreateCommentInPost(
		r.Context(),
		commentsParams,
	)
	if err != nil {
		err := fmt.Errorf("error during comment creation: ", err)
		app.respondWithError(w, r, http.StatusInternalServerError, err, "")
		return
	}
	out := output{
		ID:        comment.ID.Bytes,
		UserID:    comment.UserID.Bytes,
		PostID:    comment.PostID.Bytes,
		Content:   comment.Content,
		UpdatedAt: comment.UpdatedAt.Time,
		CreatedAt: comment.CreatedAt.Time,
	}

	app.respondWithJSON(w, r, http.StatusOK, out)
}
