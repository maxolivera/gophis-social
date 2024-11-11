// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: posts.sql

package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createPost = `-- name: CreatePost :one
INSERT INTO posts (id, created_at, updated_at, user_id, title, content, tags)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, created_at, updated_at, title, content, user_id, tags, is_deleted, version
`

type CreatePostParams struct {
	ID        pgtype.UUID
	CreatedAt pgtype.Timestamp
	UpdatedAt pgtype.Timestamp
	UserID    pgtype.UUID
	Title     string
	Content   string
	Tags      []string
}

func (q *Queries) CreatePost(ctx context.Context, arg CreatePostParams) (Post, error) {
	row := q.db.QueryRow(ctx, createPost,
		arg.ID,
		arg.CreatedAt,
		arg.UpdatedAt,
		arg.UserID,
		arg.Title,
		arg.Content,
		arg.Tags,
	)
	var i Post
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Title,
		&i.Content,
		&i.UserID,
		&i.Tags,
		&i.IsDeleted,
		&i.Version,
	)
	return i, err
}

const getPostById = `-- name: GetPostById :one
SELECT id, created_at, updated_at, title, content, user_id, tags, is_deleted, version FROM posts WHERE id = $1 AND is_deleted = false
`

func (q *Queries) GetPostById(ctx context.Context, id pgtype.UUID) (Post, error) {
	row := q.db.QueryRow(ctx, getPostById, id)
	var i Post
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Title,
		&i.Content,
		&i.UserID,
		&i.Tags,
		&i.IsDeleted,
		&i.Version,
	)
	return i, err
}

const getPostByUser = `-- name: GetPostByUser :many
SELECT id, created_at, updated_at, title, content, user_id, tags, is_deleted, version FROM posts WHERE user_id = $1 AND is_deleted = false
`

func (q *Queries) GetPostByUser(ctx context.Context, userID pgtype.UUID) ([]Post, error) {
	rows, err := q.db.Query(ctx, getPostByUser, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Post
	for rows.Next() {
		var i Post
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Title,
			&i.Content,
			&i.UserID,
			&i.Tags,
			&i.IsDeleted,
			&i.Version,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const hardDeletePostByID = `-- name: HardDeletePostByID :one
DELETE FROM posts WHERE id = $1 and version = $2 RETURNING id, created_at, updated_at, title, content, user_id, tags, is_deleted, version
`

type HardDeletePostByIDParams struct {
	ID      pgtype.UUID
	Version int32
}

func (q *Queries) HardDeletePostByID(ctx context.Context, arg HardDeletePostByIDParams) (Post, error) {
	row := q.db.QueryRow(ctx, hardDeletePostByID, arg.ID, arg.Version)
	var i Post
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Title,
		&i.Content,
		&i.UserID,
		&i.Tags,
		&i.IsDeleted,
		&i.Version,
	)
	return i, err
}

const softDeletePostByID = `-- name: SoftDeletePostByID :one
UPDATE posts
SET is_deleted = true
WHERE id = $1 and version = $2
RETURNING is_deleted
`

type SoftDeletePostByIDParams struct {
	ID      pgtype.UUID
	Version int32
}

func (q *Queries) SoftDeletePostByID(ctx context.Context, arg SoftDeletePostByIDParams) (bool, error) {
	row := q.db.QueryRow(ctx, softDeletePostByID, arg.ID, arg.Version)
	var is_deleted bool
	err := row.Scan(&is_deleted)
	return is_deleted, err
}

const updatePost = `-- name: UpdatePost :one
UPDATE posts
SET
	updated_at = $1,
	title = coalesce($4, title),
	content = coalesce($5, content),
	tags = coalesce($6, tags)
WHERE id = $2 AND is_deleted = false AND version = $3
RETURNING id, created_at, updated_at, title, content, user_id, tags, is_deleted, version
`

type UpdatePostParams struct {
	UpdatedAt pgtype.Timestamp
	ID        pgtype.UUID
	Version   int32
	Title     pgtype.Text
	Content   pgtype.Text
	Tags      []string
}

func (q *Queries) UpdatePost(ctx context.Context, arg UpdatePostParams) (Post, error) {
	row := q.db.QueryRow(ctx, updatePost,
		arg.UpdatedAt,
		arg.ID,
		arg.Version,
		arg.Title,
		arg.Content,
		arg.Tags,
	)
	var i Post
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Title,
		&i.Content,
		&i.UserID,
		&i.Tags,
		&i.IsDeleted,
		&i.Version,
	)
	return i, err
}
