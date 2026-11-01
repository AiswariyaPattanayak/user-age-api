package handler

import (
	"errors"
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"user-age-api/internal/models"
	"user-age-api/internal/service"
)

const (
	defaultPage  = 1
	defaultLimit = 20
	maxLimit     = 100
)

type UserHandler struct {
	svc      service.UserService
	validate *validator.Validate
	log      *zap.Logger
}

func NewUserHandler(svc service.UserService, log *zap.Logger) *UserHandler {
	v := validator.New()
	v.RegisterTagNameFunc(func(f reflect.StructField) string {
		name := strings.SplitN(f.Tag.Get("json"), ",", 2)[0]
		if name == "" || name == "-" {
			return f.Name
		}
		return name
	})
	return &UserHandler{svc: svc, validate: v, log: log}
}

func (h *UserHandler) Create(c *fiber.Ctx) error {
	req, err := h.bindAndValidate(c)
	if err != nil {
		return err
	}
	user, err := h.svc.Create(c.UserContext(), req)
	if err != nil {
		return h.mapError(err)
	}
	return c.Status(fiber.StatusCreated).JSON(user)
}

func (h *UserHandler) Get(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	user, err := h.svc.Get(c.UserContext(), id)
	if err != nil {
		return h.mapError(err)
	}
	return c.JSON(user)
}

func (h *UserHandler) List(c *fiber.Ctx) error {
	page := int32(c.QueryInt("page", defaultPage))
	limit := int32(c.QueryInt("limit", defaultLimit))
	if page < 1 {
		return fiber.NewError(fiber.StatusBadRequest, "page must be >= 1")
	}
	if limit < 1 || limit > maxLimit {
		return fiber.NewError(fiber.StatusBadRequest, "limit must be between 1 and "+strconv.Itoa(maxLimit))
	}

	users, total, err := h.svc.List(c.UserContext(), page, limit)
	if err != nil {
		return h.mapError(err)
	}

	// Pagination metadata is sent in headers so the body stays a plain JSON array.
	c.Set("X-Total-Count", strconv.FormatInt(total, 10))
	c.Set("X-Page", strconv.Itoa(int(page)))
	c.Set("X-Limit", strconv.Itoa(int(limit)))
	c.Set("X-Total-Pages", strconv.Itoa(int(math.Ceil(float64(total)/float64(limit)))))
	return c.JSON(users)
}

func (h *UserHandler) Update(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	req, err := h.bindAndValidate(c)
	if err != nil {
		return err
	}
	user, err := h.svc.Update(c.UserContext(), id, req)
	if err != nil {
		return h.mapError(err)
	}
	return c.JSON(user)
}

func (h *UserHandler) Delete(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	if err := h.svc.Delete(c.UserContext(), id); err != nil {
		return h.mapError(err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ---- helpers ----

func parseID(c *fiber.Ctx) (int32, error) {
	id, err := strconv.ParseInt(c.Params("id"), 10, 32)
	if err != nil || id < 1 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "id must be a positive integer")
	}
	return int32(id), nil
}

func (h *UserHandler) bindAndValidate(c *fiber.Ctx) (models.UserRequest, error) {
	var req models.UserRequest
	if err := c.BodyParser(&req); err != nil {
		return req, fiber.NewError(fiber.StatusBadRequest, "invalid JSON body")
	}
	if err := h.validate.Struct(req); err != nil {
		var verrs validator.ValidationErrors
		if errors.As(err, &verrs) {
			return req, &validationError{details: formatValidationErrors(verrs)}
		}
		return req, fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return req, nil
}

func (h *UserHandler) mapError(err error) error {
	switch {
	case errors.Is(err, models.ErrNotFound):
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	case errors.Is(err, models.ErrInvalidDOB):
		return &validationError{details: map[string]string{"dob": err.Error()}}
	default:
		return err // handled as 500 by the global error handler
	}
}

// validationError carries per-field messages to the global error handler.
type validationError struct {
	details map[string]string
}

func (e *validationError) Error() string { return "validation failed" }

func formatValidationErrors(verrs validator.ValidationErrors) map[string]string {
	out := make(map[string]string, len(verrs))
	for _, fe := range verrs {
		field := fe.Field() // json name, thanks to RegisterTagNameFunc
		switch fe.Tag() {
		case "required":
			out[field] = field + " is required"
		case "min":
			out[field] = field + " is too short"
		case "max":
			out[field] = field + " is too long"
		case "datetime":
			out[field] = field + " must be a valid date in YYYY-MM-DD format"
		default:
			out[field] = field + " is invalid"
		}
	}
	return out
}

// ErrorHandler is the global Fiber error handler: consistent JSON error bodies + proper status codes.
func ErrorHandler(log *zap.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		var ve *validationError
		if errors.As(err, &ve) {
			return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
				Error:   "validation failed",
				Details: ve.details,
			})
		}

		var fe *fiber.Error
		if errors.As(err, &fe) {
			return c.Status(fe.Code).JSON(models.ErrorResponse{Error: fe.Message})
		}

		log.Error("unhandled error",
			zap.Error(err),
			zap.String("path", c.Path()),
			zap.Any("request_id", c.Locals("requestId")),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{Error: "internal server error"})
	}
}
