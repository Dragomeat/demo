package service

import (
	"boilerplate/internal/auth/provider/local"
	"boilerplate/internal/framework"
	"boilerplate/internal/user/dao"
	"boilerplate/internal/user/storage"
	"context"
	application "github.com/debugger84/modulus-application"
	"github.com/gofrs/uuid"
	guid "github.com/google/uuid"
	"time"
)

const emailExists application.ErrorIdentifier = "emailExists"

type RegisterUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Registration struct {
	finder       *dao.UserFinder
	saver        *dao.UserSaver
	queries      *storage.Queries
	logger       framework.Logger
	authProvider *local.Provider
}

func NewRegistration(
	finder *dao.UserFinder,
	saver *dao.UserSaver,
	queries *storage.Queries,
	logger framework.Logger,
	authProvider *local.Provider,
) *Registration {
	return &Registration{finder: finder, saver: saver, queries: queries, logger: logger, authProvider: authProvider}
}

// Register returns emailExists error
func (r Registration) Register(ctx context.Context, rRequest RegisterUserRequest) (*storage.User, error) {
	if r.emailExist(ctx, rRequest.Email) {
		return nil, application.NewCommonError(emailExists, "not unique email")
	}

	id, _ := uuid.NewV6()
	request := storage.CreateUserParams{
		ID:    guid.UUID(id),
		Name:  rRequest.Name,
		Email: rRequest.Email,
	}

	user, err := r.queries.CreateUser(ctx, request)
	if err != nil {
		r.logger.Error(ctx, err.Error())
		return nil, err
	}
	account := local.LocalAccount{
		UserID:    user.ID.String(),
		Email:     user.Email,
		Password:  rRequest.Password,
		CreatedAt: time.Time{},
	}
	err = r.authProvider.Register(ctx, account)
	if err != nil {
		//@todo rollback user saving
		r.logger.Error(ctx, err.Error())
		return nil, err
	}

	return &user, nil
}

func (r Registration) emailExist(ctx context.Context, email string) bool {
	query := r.finder.CreateQuery(ctx)
	query.Email(email)
	user, _ := r.finder.OneByQuery(query)

	return user != nil
}
