package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/maxolivera/gophis-social-network/internal/database"
	"github.com/maxolivera/gophis-social-network/internal/models"
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
	in := CreatePostPayload{}
	if err := readJSON(w, r, &in); err != nil {
		err := fmt.Errorf("error during JSON decoding: %v", err)
		app.respondWithError(w, r, http.StatusInternalServerError, err, "")
		return
	}

	{ // Validate input
		if in.Title == "" || in.Content == "" {
			err := fmt.Errorf("title and content are required: %s\n", in)
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
	pgID := pgtype.UUID{Bytes: id, Valid: true}

	user := getLoggedUser(r)
	pgUserID := pgtype.UUID{Bytes: user.ID, Valid: true}

	currentTime := time.Now().UTC()
	pgTime := pgtype.Timestamp{Time: currentTime, Valid: true}

	postParams := database.CreatePostParams{
		ID:        pgID,
		UserID:    pgUserID,
		CreatedAt: pgTime,
		UpdatedAt: pgTime,
		Title:     in.Title,
		Content:   in.Content,
		Tags:      in.Tags,
	}

	// store user
	dbPost, err := app.Database.CreatePost(
		r.Context(),
		postParams,
	)
	if err != nil {
		err := fmt.Errorf("error during user creation: %v", err)
		app.respondWithError(w, r, http.StatusInternalServerError, err, "")
		return
	}

	post := models.DBPostToPost(dbPost)

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
	post := getPostFromCtx(r)

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
	post := getPostFromCtx(r)
	pgID := pgtype.UUID{
		Bytes: post.ID,
		Valid: true,
	}

	params := database.SoftDeletePostByIDParams{
		ID:      pgID,
		Version: post.Version,
	}

	deleted, err := app.Database.SoftDeletePostByID(r.Context(), params)
	if err != nil {
		switch err {
		case pgx.ErrNoRows:
			err := fmt.Errorf("post was not deleted because was not found post_id: %v", post.ID)
			app.respondWithError(w, r, http.StatusNotFound, err, "post not found")
		default:
			err := fmt.Errorf("post could not deleted: %v", err)
			app.respondWithError(w, r, http.StatusInternalServerError, err, "post could not be deleted")
		}
		return
	}
	if !deleted {
		err := fmt.Errorf("post was not deleted, post_id: %v", post.ID)
		app.respondWithError(w, r, http.StatusNotFound, err, "post not found")
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
	post := getPostFromCtx(r)
	pgID := pgtype.UUID{
		Bytes: post.ID,
		Valid: true,
	}

	params := database.HardDeletePostByIDParams{
		ID:      pgID,
		Version: post.Version,
	}

	_, err := app.Database.HardDeletePostByID(r.Context(), params)
	if err != nil {
		switch err {
		case pgx.ErrNoRows:
			err := fmt.Errorf("post not deleted, not found, post id: %v error: %v", post.ID, err)
			app.respondWithError(w, r, http.StatusNotFound, err, "post not found")
		default:
			err := fmt.Errorf("post could not deleted: %v", err)
			app.respondWithError(w, r, http.StatusInternalServerError, err, "post could not be deleted")
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
//	@Description	The post will be updated. In future will require auth.
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
	post := getPostFromCtx(r)

	in := UpdatePostPayload{}
	if err := readJSON(w, r, &in); err != nil {
		err = fmt.Errorf("error reading input parameters: %v", err)
		app.respondWithError(w, r, http.StatusInternalServerError, err, "")
	}

	currentTime := time.Now().UTC()
	pgTime := pgtype.Timestamp{Time: currentTime, Valid: true}
	pgContent := pgtype.Text{String: in.Content, Valid: len(in.Content) > 0}
	pgTitle := pgtype.Text{String: in.Title, Valid: len(in.Title) > 0}
	pgID := pgtype.UUID{Bytes: post.ID, Valid: true}

	params := database.UpdatePostParams{
		UpdatedAt: pgTime,
		ID:        pgID,
		Content:   pgContent,
		Title:     pgTitle,
		Tags:      in.Tags,
		Version:   post.Version,
	}

	newDBPost, err := app.Database.UpdatePost(r.Context(), params)
	if err != nil {
		switch err {
		// NOTE(maolivera): Considering that already retrieving with `getPostFromCtx`, is unlikely that this will happen.
		case pgx.ErrNoRows:
			app.respondWithError(w, r, http.StatusNotFound, err, "post not found")
		default:
			err := fmt.Errorf("post could not updated: %v", err)
			app.respondWithError(w, r, http.StatusInternalServerError, err, "")
		}
		return
	}
	newPost := models.DBPostToPost(newDBPost)

	app.respondWithJSON(w, r, http.StatusOK, newPost)
}
