package mapping

import (
	"github.com/preetsinghmakkar/OpenCall/internal/dtos"
	"github.com/preetsinghmakkar/OpenCall/internal/models"
)

func ToUserResponse(user *models.User) dtos.UserResponse {
	return dtos.UserResponse{
		ID:             user.ID,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		Username:       user.Username,
		Email:          user.Email,
		Role:           user.Role,
		ProfilePicture: user.ProfilePicture,
		Bio:            user.Bio,
		IsActive:       user.IsActive,
		CreatedAt:      user.CreatedAt,
	}
}

func ToUserResponseList(users []*models.User) []dtos.UserResponse {
	responses := make([]dtos.UserResponse, 0, len(users))

	for _, user := range users {
		responses = append(responses, ToUserResponse(user))
	}

	return responses
}

func ToUserFromRegisterRequest(req *dtos.RegisterUserRequest) *models.User {
	return &models.User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Username:  req.Username,
		Email:     req.Email,
		IsActive:  true,
	}
}

func ApplyUserUpdate(user *models.User, req *dtos.UpdateUserRequest) {
	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}
	if req.Bio != nil {
		user.Bio = *req.Bio
	}
	if req.ProfilePicture != nil {
		user.ProfilePicture = *req.ProfilePicture
	}
}

func ToLoginResponse(
	user *models.User,
	accessToken string,
	refreshToken string,
) dtos.LoginResponse {
	return dtos.LoginResponse{
		User:         ToUserResponse(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
}
