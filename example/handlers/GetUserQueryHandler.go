package handlers

import (
	"context"
	"fmt"

	"github.com/kmdeveloping/go-cqrs/example/queries"
	"github.com/kmdeveloping/go-cqrs/query"
)

// GetUserQueryHandler demonstrates dependency injection with auto-registration
type GetUserQueryHandler struct {
	// Dependencies auto-injected using the 'inject' tag
	UserRepo UserRepository `inject:""`
	Logger   Logger         `inject:""`
}

// Ensure handler implements the interface
var _ query.IQueryHandler[queries.GetUserQuery, queries.GetUserQueryResponse] = (*GetUserQueryHandler)(nil)

func (h *GetUserQueryHandler) Handle(ctx context.Context, q queries.GetUserQuery) (queries.GetUserQueryResponse, error) {
	h.Logger.Infof("Processing GetUserQuery for user ID: %d", q.ID)

	// Get user using injected repository
	user, err := h.UserRepo.GetByID(q.ID)
	if err != nil {
		h.Logger.Error("User not found")
		return queries.GetUserQueryResponse{}, fmt.Errorf("failed to get user: %w", err)
	}

	h.Logger.Info("User retrieved successfully")

	// Return response
	return queries.GetUserQueryResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}, nil
}
