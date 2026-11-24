package auth

import (
	"context"
	"errors"
)

type TestUser struct {
	Username string
	Password string
}

var DevelopmentTestUsers = []TestUser{
	{Username: "alex", Password: "UploadHub-Test-2026-Alex!"},
	{Username: "morgan", Password: "UploadHub-Test-2026-Morgan!"},
	{Username: "taylor", Password: "UploadHub-Test-2026-Taylor!"},
}

func SeedDevelopmentUsers(ctx context.Context, service *Service) error {
	for _, testUser := range DevelopmentTestUsers {
		_, err := service.CreateUser(ctx, testUser.Username, testUser.Password)
		if err != nil && !errors.Is(err, ErrUsernameTaken) {
			return err
		}
	}

	return nil
}
