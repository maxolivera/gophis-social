package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/maxolivera/gophis-social-network/internal/database"
	"github.com/maxolivera/gophis-social-network/internal/storage"
	"github.com/maxolivera/gophis-social-network/internal/storage/models"
	"golang.org/x/crypto/bcrypt"
)

type PostgresUserRepository struct {
	p *pgxpool.Pool
}

// Fetch a userhashing by username
func (r PostgresUserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(ctx, storage.QueryTimeDuration)
	defer cancel()

	q := database.New(r.p)
	dbUser, err := q.GetUserByUsername(ctx, username)
	if err != nil {
		switch err {
		case pgx.ErrNoRows:
			return nil, storage.ErrNoRows
		default:
			return nil, err
		}
	}

	user := models.DBUserWithRoleToUser(dbUser)
	return user, nil
}

// Fetch a user by email
func (r PostgresUserRepository) GetByEmailAndPassword(ctx context.Context, email, pass string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(ctx, storage.QueryTimeDuration)
	defer cancel()

	q := database.New(r.p)
	dbUser, err := q.GetUserByEmail(ctx, email)
	if err != nil {
		switch err {
		case pgx.ErrNoRows:
			return nil, storage.ErrNoRows
		default:
			return nil, err
		}
	}

	// compare password
	err = bcrypt.CompareHashAndPassword(dbUser.Password, []byte(pass))
	if err != nil {
		err := fmt.Errorf("error when comparing passwords: %v", err)
		return nil, err
	}

	user := models.DBUserToUser(dbUser)
	return user, nil
}

// Stores a user (no transaction)
func (r PostgresUserRepository) Create(ctx context.Context, u *models.UserWithPassword) error {
	q := database.New(r.p)

	// hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return q.CreateUser(ctx, database.CreateUserParams{
		ID:        pgtype.UUID{Bytes: u.User.ID, Valid: true},
		CreatedAt: pgtype.Timestamp{Time: u.User.CreatedAt, Valid: true},
		UpdatedAt: pgtype.Timestamp{Time: u.User.UpdatedAt, Valid: true},
		Username:  u.User.Username,
		Email:     u.User.Email,
		Password:  hashed,
	})
}

// Stores a user (transcation)
func (r PostgresUserRepository) createWithTx(ctx context.Context, u *models.UserWithPassword, tx pgx.Tx) error {
	q := database.New(r.p)
	qtx := q.WithTx(tx)

	// hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if err = qtx.CreateUser(ctx, database.CreateUserParams{
		ID:        pgtype.UUID{Bytes: u.User.ID, Valid: true},
		CreatedAt: pgtype.Timestamp{Time: u.User.CreatedAt, Valid: true},
		UpdatedAt: pgtype.Timestamp{Time: u.User.UpdatedAt, Valid: true},
		Username:  u.User.Username,
		Email:     u.User.Email,
		Password:  hashed,
	}); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			switch pgErr.ConstraintName {
			case "users_username_key":
				return storage.ErrUsernameUnavailable
			case "users_email_key":
				return storage.ErrEmailUnavailable
			default:
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

// Stores a user and the invitation
func (r PostgresUserRepository) CreateAndInvite(ctx context.Context, user *models.UserWithPassword, token []byte, invitationExp time.Duration) error {
	return withTx(r.p, ctx, func(tx pgx.Tx) error {
		// Create transaction
		q := database.New(r.p)
		qtx := q.WithTx(tx)

		if err := r.createWithTx(ctx, user, tx); err != nil {
			return err
		}

		return qtx.CreateInvitation(ctx, database.CreateInvitationParams{
			UserID:    pgtype.UUID{Bytes: user.User.ID, Valid: true},
			Token:     token,
			ExpiresAt: pgtype.Timestamp{Time: time.Now().UTC().Add(invitationExp), Valid: true},
		})
	})
}

// Activates a user and deletes the invitation
func (r PostgresUserRepository) Activate(ctx context.Context, token []byte) (*models.User, error) {
	var user *models.User

	if err := withTx(r.p, ctx, func(tx pgx.Tx) error {
		q := database.New(r.p)
		qtx := q.WithTx(tx)

		// 1. Validate token
		id, err := qtx.GetInvitation(ctx, database.GetInvitationParams{
			Token:     token,
			ExpiresAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
		})
		if err != nil {
			if err == pgx.ErrNoRows {
				return storage.ErrNoToken
			} else {
				return err
			}
		}

		// 2. Activate user
		dbUser, err := qtx.ActivateUser(ctx, id)
		if err != nil {
			if err == pgx.ErrNoRows {
				return storage.ErrNoUser
			} else {
				return err
			}
		}

		// 3. Delete token from DB
		if err = qtx.DeleteToken(ctx, token); err != nil {
			if err == pgx.ErrNoRows {
				return storage.ErrNoToken
			} else {
				return err
			}
		}

		user = models.DBUserToUser(dbUser)

		return nil
	}); err != nil {
		return nil, err
	}

	return user, nil
}

// Mark a user as deleted
func (r PostgresUserRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	q := database.New(r.p)

	return q.SoftDeleteUserByID(ctx, pgtype.UUID{Bytes: id, Valid: true})
}

// Deletes a user
func (r PostgresUserRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	q := database.New(r.p)

	return q.HardDeleteUserByID(ctx, pgtype.UUID{Bytes: id, Valid: true})
}

// Updates a user. The `u` parameter may contain empty fields, which mean they will not change.
func (r PostgresUserRepository) Update(ctx context.Context, u *models.UserWithPassword) (*models.User, error) {
	q := database.New(r.p)
	currentTime := time.Now().UTC()
	var pgPassword []byte
	if u.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(u.Password), 14)
		if err != nil {
			return nil, err
		}
		pgPassword = hashed
	} else {
		pgPassword = nil
	}

	dbUser, err := q.UpdateUser(ctx, database.UpdateUserParams{
		UpdatedAt: pgtype.Timestamp{Time: currentTime, Valid: true},
		ID:        pgtype.UUID{Bytes: u.User.ID, Valid: true},
		Username:  pgtype.Text{String: u.User.Username, Valid: len(u.User.Username) > 0 || len(u.User.Username) <= 100},
		Email:     pgtype.Text{String: u.User.Email, Valid: len(u.User.Email) > 0 || len(u.User.Email) <= 255},
		FirstName: pgtype.Text{String: u.User.FirstName, Valid: len(u.User.FirstName) > 0 || len(u.User.FirstName) <= 100},
		LastName:  pgtype.Text{String: u.User.LastName, Valid: len(u.User.LastName) > 0 || len(u.User.LastName) <= 100},
		Password:  pgPassword,
	})
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			switch pgErr.ConstraintName {
			case "users_username_key":
				return nil, storage.ErrUsernameUnavailable
			case "users_email_key":
				return nil, storage.ErrEmailUnavailable
			default:
				return nil, err
			}
		}
		return nil, err
	}

	user := models.DBUserToUser(dbUser)

	return user, nil
}
