package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/maxolivera/gophis-social-network/internal/database"
	"github.com/maxolivera/gophis-social-network/internal/storage"
	"github.com/maxolivera/gophis-social-network/internal/storage/models"
)

type PostgresCommentRepository struct {
	p *pgxpool.Pool
}

func (r *PostgresCommentRepository) GetByPostID(ctx context.Context, id uuid.UUID) ([]*models.Comment, error) {
	q := database.New(r.p)
	dbComments, err := q.GetCommentsByPost(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		switch err {
		case pgx.ErrNoRows:
			return nil, storage.ErrNoRows
		default:
			return nil, err
		}
	}

	comments := models.DBCommentsWithUser(dbComments)

	return comments, nil
}

func (r *PostgresCommentRepository) Create(ctx context.Context, comment *models.Comment) error {
	q := database.New(r.p)
	return q.CreateCommentInPost(ctx, database.CreateCommentInPostParams{
		ID:        pgtype.UUID{Bytes: comment.ID, Valid: true},
		UserID:    pgtype.UUID{Bytes: comment.User.ID, Valid: true},
		PostID:    pgtype.UUID{Bytes: comment.PostID, Valid: true},
		CreatedAt: pgtype.Timestamp{Time: comment.CreatedAt, Valid: true},
		UpdatedAt: pgtype.Timestamp{Time: comment.UpdatedAt, Valid: true},
		Content:   comment.Content,
	})
}
