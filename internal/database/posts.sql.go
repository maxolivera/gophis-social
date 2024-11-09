// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: posts.sql

package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createPosts = `-- name: CreatePosts :one
INSERT INTO posts (id, created_at, updated_at, user_id, title, content, tags)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, created_at, title
`

type CreatePostsParams struct {
	ID        pgtype.UUID
	CreatedAt pgtype.Timestamp
	UpdatedAt pgtype.Timestamp
	UserID    pgtype.UUID
	Title     string
	Content   string
	Tags      []string
}

type CreatePostsRow struct {
	ID        pgtype.UUID
	CreatedAt pgtype.Timestamp
	Title     string
}

func (q *Queries) CreatePosts(ctx context.Context, arg CreatePostsParams) (CreatePostsRow, error) {
	row := q.db.QueryRow(ctx, createPosts,
		arg.ID,
		arg.CreatedAt,
		arg.UpdatedAt,
		arg.UserID,
		arg.Title,
		arg.Content,
		arg.Tags,
	)
	var i CreatePostsRow
	err := row.Scan(&i.ID, &i.CreatedAt, &i.Title)
	return i, err
}

const getPostById = `-- name: GetPostById :one
SELECT id, created_at, updated_at, title, content, user_id, tags FROM posts WHERE id = $1
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
	)
	return i, err
}

const getPostByUser = `-- name: GetPostByUser :many
SELECT id, created_at, updated_at, title, content, user_id, tags FROM posts WHERE user_id = $1
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
