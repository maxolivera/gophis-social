package api

import (
	"context"
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

func (app *Application) middlewarePostContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		idStr := r.PathValue("postID")
		if idStr == "" {
			err := fmt.Errorf("post_id not provided")
			// TODO(maolivera): maybe another message?
			respondWithError(w, r, http.StatusBadRequest, err, err.Error())
			return
		}

		id, err := uuid.Parse(idStr)
		if err != nil {
			err := fmt.Errorf("post_id not valid: %v", err)
			respondWithError(w, r, http.StatusBadRequest, err, "post not found")
			return
		}

		pgID := pgtype.UUID{
			Bytes: id,
			Valid: true,
		}

		dbPost, err := app.Database.GetPostById(
			ctx,
			pgID,
		)

		if err != nil {
			switch err {
			case pgx.ErrNoRows:
				err := fmt.Errorf("post_id not found: %v", err)
				respondWithError(w, r, http.StatusNotFound, err, "post not found")
			default:
				respondWithError(w, r, http.StatusInternalServerError, err, "")
			}
			return
		}
		dbComments, err := app.Database.GetCommentsByPost(r.Context(), dbPost.ID)
		if err != nil {
			switch err {
			case pgx.ErrNoRows:
			// TODO(maolivera): maybe add logging? but seems uncessary to add logs if no comments found
			default:
				respondWithError(w, r, http.StatusInternalServerError, err, "")
			}
			return
		}
		post := models.DBPostToPost(dbPost)
		comments := models.DBCommentsWithUser(dbComments)
		post.Comments = comments

		ctx = context.WithValue(ctx, "post", &post)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getPostFromCtx(r *http.Request) *models.Post {
	return r.Context().Value("post").(*models.Post)
}

func (app *Application) handlerCreatePost(w http.ResponseWriter, r *http.Request) {
	// types for JSON's input and output
	type input struct {
		Title   string    `json:"title"`
		Content string    `json:"content"`
		UserID  uuid.UUID `json:"user_id"`
		Tags    []string  `json:"tags"`
	}
	type output struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		Title     string    `json:"title"`
	}
	in := input{}

	if err := readJSON(w, r, &in); err != nil {
		err := fmt.Errorf("error during JSON decoding: %v", err)
		respondWithError(w, r, http.StatusInternalServerError, err, "")
		return
	}

	// check if some parameter was null
	if in.Title == "" || in.Content == "" {
		err := fmt.Errorf("title and content are required: %s\n", in)
		respondWithError(w, r, http.StatusBadRequest, err, "something is missing")
		return
	}
	if len(in.Title) > MAX_TITLE_LENGTH {
		err := fmt.Errorf("title is too long, max is %d vs. current %d", MAX_TITLE_LENGTH, len(in.Title))
		respondWithError(w, r, http.StatusBadRequest, err, "title is too long")
		return
	}
	if len(in.Content) > MAX_CONTENT_LENGTH {
		err := fmt.Errorf("content is too long, max is %d vs. current %d", MAX_CONTENT_LENGTH, len(in.Content))
		respondWithError(w, r, http.StatusBadRequest, err, "content is too long")
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
	post, err := app.Database.CreatePost(
		r.Context(),
		postParams,
	)
	if err != nil {
		err := fmt.Errorf("error during user creation: %v", err)
		respondWithError(w, r, http.StatusInternalServerError, err, "")
		return
	}

	// Marshal response
	out := output{}
	out.CreatedAt = post.CreatedAt.Time
	out.Title = post.Title
	out.ID = post.ID.Bytes

	// Send response
	respondWithJSON(w, http.StatusOK, out)
}

func (app *Application) handlerGetPost(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)

	respondWithJSON(w, http.StatusOK, post)
}

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
			respondWithError(w, r, http.StatusNotFound, err, "post not found")
		default:
			err := fmt.Errorf("post could not deleted: %v", err)
			respondWithError(w, r, http.StatusInternalServerError, err, "post could not be deleted")
		}
		return
	}
	if !deleted {
		err := fmt.Errorf("post was not deleted, post_id: %v", post.ID)
		respondWithError(w, r, http.StatusNotFound, err, "post not found")
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}

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
			respondWithError(w, r, http.StatusNotFound, err, "post not found")
		default:
			err := fmt.Errorf("post could not deleted: %v", err)
			respondWithError(w, r, http.StatusInternalServerError, err, "post could not be deleted")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, nil)
}

func (app *Application) handlerUpdatePost(w http.ResponseWriter, r *http.Request) {
	type input struct {
		Content string   `json:"content,omitempty"`
		Title   string   `json:"title,omitempty"`
		Tags    []string `json:"tags,omitempty"`
	}
	post := getPostFromCtx(r)

	in := input{}
	if err := readJSON(w, r, &in); err != nil {
		err = fmt.Errorf("error reading input parameters: %v", err)
		respondWithError(w, r, http.StatusInternalServerError, err, "")
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
		case pgx.ErrNoRows:
			respondWithError(w, r, http.StatusNotFound, err, "post not found")
		default:
			err := fmt.Errorf("post could not updated: %v", err)
			respondWithError(w, r, http.StatusInternalServerError, err, "")
		}
		return
	}
	newPost := models.DBPostToPost(newDBPost)

	respondWithJSON(w, http.StatusOK, newPost)
}
