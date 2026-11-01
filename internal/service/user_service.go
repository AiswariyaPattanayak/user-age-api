package service

import (
	"context"
	"time"

	"go.uber.org/zap"

	"user-age-api/db/sqlc"
	"user-age-api/internal/models"
	"user-age-api/internal/repository"
)

type UserService interface {
	Create(ctx context.Context, req models.UserRequest) (models.UserResponse, error)
	Get(ctx context.Context, id int32) (models.UserWithAgeResponse, error)
	List(ctx context.Context, page, limit int32) ([]models.UserWithAgeResponse, int64, error)
	Update(ctx context.Context, id int32, req models.UserRequest) (models.UserResponse, error)
	Delete(ctx context.Context, id int32) error
}

type userService struct {
	repo  repository.UserRepository
	log   *zap.Logger
	nowFn func() time.Time
}

func NewUserService(repo repository.UserRepository, log *zap.Logger) UserService {
	return &userService{repo: repo, log: log, nowFn: time.Now}
}

// parseDOB parses YYYY-MM-DD and rejects dates in the future.
func (s *userService) parseDOB(raw string) (time.Time, error) {
	dob, err := time.Parse(models.DateLayout, raw)
	if err != nil {
		return time.Time{}, models.ErrInvalidDOB
	}
	now := s.nowFn().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	if dob.After(today) {
		return time.Time{}, models.ErrInvalidDOB
	}
	return dob, nil
}

func toResponse(u sqlc.User) models.UserResponse {
	return models.UserResponse{ID: u.ID, Name: u.Name, DOB: u.Dob.Format(models.DateLayout)}
}

func (s *userService) toResponseWithAge(u sqlc.User) models.UserWithAgeResponse {
	return models.UserWithAgeResponse{
		ID:   u.ID,
		Name: u.Name,
		DOB:  u.Dob.Format(models.DateLayout),
		Age:  CalculateAge(u.Dob, s.nowFn().UTC()),
	}
}

func (s *userService) Create(ctx context.Context, req models.UserRequest) (models.UserResponse, error) {
	dob, err := s.parseDOB(req.DOB)
	if err != nil {
		return models.UserResponse{}, err
	}
	u, err := s.repo.Create(ctx, req.Name, dob)
	if err != nil {
		s.log.Error("failed to create user", zap.Error(err))
		return models.UserResponse{}, err
	}
	s.log.Info("user created", zap.Int32("user_id", u.ID))
	return toResponse(u), nil
}

func (s *userService) Get(ctx context.Context, id int32) (models.UserWithAgeResponse, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return models.UserWithAgeResponse{}, err
	}
	s.log.Info("user fetched", zap.Int32("user_id", id))
	return s.toResponseWithAge(u), nil
}

func (s *userService) List(ctx context.Context, page, limit int32) ([]models.UserWithAgeResponse, int64, error) {
	offset := (page - 1) * limit
	users, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		s.log.Error("failed to list users", zap.Error(err))
		return nil, 0, err
	}
	total, err := s.repo.Count(ctx)
	if err != nil {
		s.log.Error("failed to count users", zap.Error(err))
		return nil, 0, err
	}
	out := make([]models.UserWithAgeResponse, 0, len(users))
	for _, u := range users {
		out = append(out, s.toResponseWithAge(u))
	}
	s.log.Info("users listed", zap.Int("returned", len(out)), zap.Int32("page", page), zap.Int32("limit", limit))
	return out, total, nil
}

func (s *userService) Update(ctx context.Context, id int32, req models.UserRequest) (models.UserResponse, error) {
	dob, err := s.parseDOB(req.DOB)
	if err != nil {
		return models.UserResponse{}, err
	}
	u, err := s.repo.Update(ctx, id, req.Name, dob)
	if err != nil {
		if err != models.ErrNotFound {
			s.log.Error("failed to update user", zap.Int32("user_id", id), zap.Error(err))
		}
		return models.UserResponse{}, err
	}
	s.log.Info("user updated", zap.Int32("user_id", id))
	return toResponse(u), nil
}

func (s *userService) Delete(ctx context.Context, id int32) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		if err != models.ErrNotFound {
			s.log.Error("failed to delete user", zap.Int32("user_id", id), zap.Error(err))
		}
		return err
	}
	s.log.Info("user deleted", zap.Int32("user_id", id))
	return nil
}
