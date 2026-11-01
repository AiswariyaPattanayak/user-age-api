package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"user-age-api/db/sqlc"
	"user-age-api/internal/models"
)

// UserRepository abstracts persistence so the service layer can be tested with fakes.
type UserRepository interface {
	Create(ctx context.Context, name string, dob time.Time) (sqlc.User, error)
	GetByID(ctx context.Context, id int32) (sqlc.User, error)
	List(ctx context.Context, limit, offset int32) ([]sqlc.User, error)
	Count(ctx context.Context) (int64, error)
	Update(ctx context.Context, id int32, name string, dob time.Time) (sqlc.User, error)
	Delete(ctx context.Context, id int32) error
}

type userRepository struct {
	q sqlc.Querier
}

func NewUserRepository(q sqlc.Querier) UserRepository {
	return &userRepository{q: q}
}

func (r *userRepository) Create(ctx context.Context, name string, dob time.Time) (sqlc.User, error) {
	return r.q.CreateUser(ctx, sqlc.CreateUserParams{Name: name, Dob: dob})
}

func (r *userRepository) GetByID(ctx context.Context, id int32) (sqlc.User, error) {
	u, err := r.q.GetUser(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return sqlc.User{}, models.ErrNotFound
	}
	return u, err
}

func (r *userRepository) List(ctx context.Context, limit, offset int32) ([]sqlc.User, error) {
	return r.q.ListUsers(ctx, sqlc.ListUsersParams{Limit: limit, Offset: offset})
}

func (r *userRepository) Count(ctx context.Context) (int64, error) {
	return r.q.CountUsers(ctx)
}

func (r *userRepository) Update(ctx context.Context, id int32, name string, dob time.Time) (sqlc.User, error) {
	u, err := r.q.UpdateUser(ctx, sqlc.UpdateUserParams{ID: id, Name: name, Dob: dob})
	if errors.Is(err, sql.ErrNoRows) {
		return sqlc.User{}, models.ErrNotFound
	}
	return u, err
}

func (r *userRepository) Delete(ctx context.Context, id int32) error {
	rows, err := r.q.DeleteUser(ctx, id)
	if err != nil {
		return err
	}
	if rows == 0 {
		return models.ErrNotFound
	}
	return nil
}
