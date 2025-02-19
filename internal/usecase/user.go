package usecase

import (
	"fmt"
	"tourism-backend/internal/entity"
	"tourism-backend/internal/usecase/repo"
	"tourism-backend/utils"
)

type UserUseCase struct {
	repo *repo.UserRepo
}

// NewTourismUseCase -.
func NewUserUseCase(r *repo.UserRepo) *UserUseCase {
	return &UserUseCase{
		repo: r,
	}
}

func (u *UserUseCase) LoginUser(user *entity.LoginUserDTO) (string, error) {
	userFromRepo, err := u.repo.LoginUser(user)
	if err != nil {
		return "", fmt.Errorf("User From Repo: %w", err)
	}
	if !utils.CheckPassword(userFromRepo.Password, user.Password) {
		return "", fmt.Errorf("Check Password: %w", err)
	}
	token, err := utils.GenerateJWT(userFromRepo.ID, userFromRepo.Role)
	if err != nil {
		return "", fmt.Errorf("Generate JWT: %w", err)
	}
	return token, nil
}

func (u *UserUseCase) RegisterUser(user *entity.User) (*entity.User, error) {
	user, err := u.repo.RegisterUser(user)
	if err != nil {
		return nil, fmt.Errorf("register user: %w", err)
	}
	return user, nil
}
