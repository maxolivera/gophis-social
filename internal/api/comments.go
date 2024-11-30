package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/maxolivera/gophis-social-network/internal/storage/models"
)

type CreateCommentPayload struct {
	Content string `json:"content"`
}

// Create Comment godoc
//
//	@Summary		Creates a comment
//	@Description	Logged user will publicate a comment on a post
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			Payload	body		CreateCommentPayload	true	"Comment"
//	@Param			PostID	path		string					true	"Post ID"
//	@Success		200		{object}	models.Comment
//	@Failure		500		{object}	error	"Something went wrong on the server"
//	@Failure		401		{object}	error	"User not logged in"
//	@Failure		404		{object}	error	"User or post not found"
//	@Failure		400		{object}	error	"Some parameter was either not provided or is invalid (e.g. content too long)"
//	@Security		ApiKeyAuth
//	@Router			/posts/{PostID}/comment [post]
func (app *Application) handlerCreateComment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getLoggedUser(r)
	post := getPost(r)

	in := CreateCommentPayload{}
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

	// create comment
	id := uuid.New()
	currentTime := time.Now().UTC()
	comment := &models.Comment{
		ID:        id,
		User:      user,
		PostID:    post.ID,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
		Content:   in.Content,
	}

	if err := app.Storage.Comments.Create(ctx, comment); err != nil {
		err := fmt.Errorf("error during comment creation: %v", err)
		app.respondWithError(w, r, http.StatusInternalServerError, err, "")
		return
	}

	app.respondWithJSON(w, r, http.StatusOK, comment)
}
