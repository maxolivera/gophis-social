package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/maxolivera/gophis-social-network/internal/database"
	"github.com/maxolivera/gophis-social-network/internal/storage"
	"github.com/maxolivera/gophis-social-network/internal/storage/models"
)

type PostgresRoleRepository struct {
	p *pgxpool.Pool
}

func (r PostgresRoleRepository) GetByName(ctx context.Context, name string) (*models.ReducedRole, error) {
	q := database.New(r.p)

	dbRole, err := q.GetRoleByName(ctx, name)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, storage.ErrNoRows
		}
		return nil, err
	}

	return &models.ReducedRole{
		Name:  models.RoleType(dbRole.Name),
		Level: int(dbRole.Level),
	}, nil
}
