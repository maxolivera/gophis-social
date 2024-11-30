package api

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/maxolivera/gophis-social-network/internal/storage"
	"github.com/maxolivera/gophis-social-network/internal/storage/models"
)

const MAX_TITLE_LENGTH = 200
const MAX_CONTENT_LENGTH = 1000

type CreatePostPayload struct {
	Title   string   `json:"title"`
	Content string   `json:"content"`
	Tags    []string `json:"tags"`
}

// Create Post godoc
//
//	@Summary		Creates a post
//	@Description	Logged user will publicate a post
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			Payload	body		CreatePostPayload	true	"Post content"
//	@Success		200		{object}	models.Post
//	@Failure		500		{object}	error	"Something went wrong on the server"
//	@Failure		400		{object}	error	"Some parameter was either not provided or is invalid (e.g. title too long)"
//	@Security		ApiKeyAuth
//	@Router			/posts [post]
func (app *Application) handlerCreatePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	in := CreatePostPayload{}
	if err := readJSON(w, r, &in); err != nil {
		err := fmt.Errorf("error during JSON decoding: %v", err)
		app.respondWithError(w, r, http.StatusInternalServerError, err, "")
		return
	}

	{ // Validate input
		if in.Title == "" || in.Content == "" {
			err := fmt.Errorf("title and content are required: %s", in)
			app.respondWithError(w, r, http.StatusBadRequest, err, "something is missing")
			return
		}
		if len(in.Title) > MAX_TITLE_LENGTH {
			err := fmt.Errorf("title is too long, max is %d vs. current %d", MAX_TITLE_LENGTH, len(in.Title))
			app.respondWithError(w, r, http.StatusBadRequest, err, "title is too long")
			return
		}
		if len(in.Content) > MAX_CONTENT_LENGTH {
			err := fmt.Errorf("content is too long, max is %d vs. current %d", MAX_CONTENT_LENGTH, len(in.Content))
			app.respondWithError(w, r, http.StatusBadRequest, err, "content is too long")
			return
		}
	}

	// create post
	id := uuid.New()
	user := getLoggedUser(r)
	currentTime := time.Now().UTC()

	post := &models.Post{
		ID:        id,
		UserID:    user.ID,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
		Title:     in.Title,
		Content:   in.Content,
		Tags:      in.Tags,
	}
	// store user
	err := app.Storage.Posts.Create(ctx, post)
	if err != nil {
		app.respondWithError(w, r, http.StatusInternalServerError, err, "")
	}

	// Send response
	app.respondWithJSON(w, r, http.StatusOK, post)
}

// Get Post godoc
//
//	@Summary		Fetch a post
//	@Description	Fetch a post
//	@Tags			posts
//	@Produce		json
//	@Param			postID	path		string	true	"Post ID"
//	@Success		200		{object}	models.Post
//	@Failure		404		{object}	error	"Post not found"
//	@Failure		500		{object}	error	"Something went wrong on the server"
//	@Security		ApiKeyAuth
//	@Router			/posts/{postID} [get]
func (app *Application) handlerGetPost(w http.ResponseWriter, r *http.Request) {
	post := getPost(r)

	app.respondWithJSON(w, r, http.StatusOK, post)
}

// Soft Delete Post godoc
//
//	@Summary		Soft Deletes a Post
//	@Description	The post will be marked as "deleted" on the database, it will not appear in any feed nor it can be accessed, but it will not be deleted from the database
//	@Tags			posts
//	@Produce		json
//	@Param			postID	path	uuid	true	"Post ID"
//	@Success		204		"The post was deleted"
//	@Failure		404		{object}	error	"Post not found"
//	@Failure		500		{object}	error	"Something went wrong on the server"
//	@Security		ApiKeyAuth
//	@Router			/posts/{postID} [delete]
func (app *Application) handlerSoftDeletePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	post := getPost(r)

	err := app.Storage.Posts.SoftDelete(ctx, post)
	if err != nil {
		switch err {
		case storage.ErrNoRows:
			err := errors.New("post not found")
			app.respondWithError(w, r, http.StatusNotFound, err, err.Error())
		default:
			app.respondWithError(w, r, http.StatusInternalServerError, err, "")
		}
		return
	}

	app.respondWithJSON(w, r, http.StatusNoContent, nil)
}

// Hard Delete Post godoc
//
//	@Summary		Hard Deletes a Post
//	@Description	The post will be deleted. Only for admins.
//	@Tags			posts, admin
//	@Produce		json
//	@Param			postID	path	uuid	true	"Post ID"
//	@Success		204		"The post was deleted"
//	@Failure		404		{object}	error	"Post not found"
//	@Failure		500		{object}	error	"Something went wrong on the server"
//	@Security		ApiKeyAuth
//	@Router			/posts/{postID}/hard [delete]
func (app *Application) handlerHardDeletePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	post := getPost(r)

	err := app.Storage.Posts.HardDelete(ctx, post)
	if err != nil {
		switch err {
		case storage.ErrNoRows:
			err := errors.New("post not found")
			app.respondWithError(w, r, http.StatusNotFound, err, err.Error())
		default:
			app.respondWithError(w, r, http.StatusInternalServerError, err, "")
		}
		return
	}

	app.respondWithJSON(w, r, http.StatusNoContent, nil)
}

type UpdatePostPayload struct {
	Title   string   `json:"title,omitempty"`
	Content string   `json:"content,omitempty"`
	Tags    []string `json:"tags,omitempty"`
}

// Update Post godoc
//
//	@Summary		Updates a Post
//	@Description	Updates a Post.
//	@Tags			posts
//	@Produce		json
//	@Param			postID	path		string				true	"Post ID"
//	@Param			Payload	body		UpdatePostPayload	true	"Updated post payload"
//	@Success		200		{object}	models.Post			"New Post"
//	@Failure		404		{object}	error				"Post not found"
//	@Failure		500		{object}	error				"Something went wrong on the server"
//	@Security		ApiKeyAuth
//	@Router			/posts/{postID} [patch]
func (app *Application) handlerUpdatePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	post := getPost(r)

	in := UpdatePostPayload{}
	if err := readJSON(w, r, &in); err != nil {
		err = fmt.Errorf("error reading input parameters: %v", err)
		app.respondWithError(w, r, http.StatusInternalServerError, err, "")
	}

	newPost := &models.Post{
		Title:   in.Title,
		Content: in.Content,
		Tags:    in.Tags,
		Version: post.Version,
	}

	updatedPost, err := app.Storage.Posts.Update(ctx, newPost)
	if err != nil {
		switch err {
		case storage.ErrNoRows:
			err = fmt.Errorf("post with id: %v not found", post.ID)
			app.respondWithError(w, r, http.StatusNotFound, err, "post not found")
		default:
			app.respondWithError(w, r, http.StatusInternalServerError, err, "")
		}
		return
	}

	app.respondWithJSON(w, r, http.StatusOK, updatedPost)
}
