package services

import (
	"errors"
	"time"

	"github.com/preetsinghmakkar/OpenCall/internal/dtos"
	"github.com/preetsinghmakkar/OpenCall/internal/mapping"
	"github.com/preetsinghmakkar/OpenCall/internal/repositories"
	"github.com/preetsinghmakkar/OpenCall/internal/utils"
)

type AuthService struct {
	userRepo         *repositories.UserRepository
	refreshTokenRepo *repositories.RefreshTokenRepository
	jwtSecret        string
}

func NewAuthService(
	userRepo *repositories.UserRepository,
	refreshTokenRepo *repositories.RefreshTokenRepository,
	jwtSecret string,
) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtSecret:        jwtSecret,
	}
}

func (s *AuthService) Login(
	req *dtos.LoginRequest,
) (*dtos.LoginResponse, error) {

	user, err := s.userRepo.FindByEmailOrUsername(req.Identifier)
	if err != nil || !user.IsActive {
		return nil, errors.New("invalid credentials")
	}

	if err := utils.ComparePassword(user.PasswordHash, req.Password); err != nil {
		return nil, errors.New("invalid credentials")
	}

	accessToken, err := utils.GenerateAccessToken(
		user.ID,
		user.Role,
		s.jwtSecret,
	)
	if err != nil {
		return nil, err
	}

	refreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	refreshTokenHash := utils.HashRefreshToken(refreshToken)

	err = s.refreshTokenRepo.Create(
		user.ID,
		refreshTokenHash,
		time.Now().Add(10*24*time.Hour),
	)
	if err != nil {
		return nil, err
	}

	return &dtos.LoginResponse{
		User:         mapping.ToUserResponse(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    utils.AccessTokenTTLSeconds(),
	}, nil
}

func (s *AuthService) RefreshAccessToken(
	req *dtos.RefreshTokenRequest,
) (*dtos.RefreshTokenResponse, error) {

	tokenHash := utils.HashRefreshToken(req.RefreshToken)

	storedToken, err := s.refreshTokenRepo.FindByTokenHash(tokenHash)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// 🚨 REUSE DETECTION
	if storedToken.RevokedAt != nil {
		_ = s.refreshTokenRepo.RevokeAllForUser(storedToken.UserID)
		return nil, errors.New("refresh token reuse detected")
	}

	if time.Now().After(storedToken.ExpiresAt) {
		return nil, errors.New("refresh token expired")
	}

	user, err := s.userRepo.FindByID(storedToken.UserID)
	if err != nil || !user.IsActive {
		return nil, errors.New("user not allowed")
	}

	// Rotate
	if err := s.refreshTokenRepo.Revoke(storedToken.ID); err != nil {
		return nil, err
	}

	newRefreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	newHash := utils.HashRefreshToken(newRefreshToken)

	err = s.refreshTokenRepo.Create(
		user.ID,
		newHash,
		time.Now().Add(10*24*time.Hour),
	)
	if err != nil {
		return nil, err
	}

	accessToken, err := utils.GenerateAccessToken(
		user.ID,
		user.Role,
		s.jwtSecret,
	)
	if err != nil {
		return nil, err
	}

	return &dtos.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    utils.AccessTokenTTLSeconds(),
	}, nil
}
