package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/preetsinghmakkar/OpenCall/internal/dtos"
	"github.com/preetsinghmakkar/OpenCall/internal/services"
)

type User struct {
	userService *services.User
}

func NewUserHandler(userService *services.User) *User {
	return &User{
		userService: userService,
	}
}

func (h *User) CreateUser(ctx *gin.Context) {
	var registerUserRequest dtos.RegisterUserRequest

	if err := ctx.ShouldBindJSON(&registerUserRequest); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			out := make(map[string]string)
			for _, fe := range ve {
				out[fe.Field()] = msgForTag(fe)
			}
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"errors": out})

			return
		}
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	createUserResponse, signupError := h.userService.CreateUser(&registerUserRequest)
	if signupError != nil {
		ctx.AbortWithStatusJSON(signupError.Status, gin.H{
			"code":    signupError.Code,
			"message": signupError.Message,
		})

		return
	}

	ctx.JSON(http.StatusCreated, createUserResponse)
}

func msgForTag(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "min":
		return fmt.Sprintf("Minimum length is %s", fe.Param())
	case "custom_password":
		return "Password must be at least 8 characters long and include uppercase, lowercase, number, and special character"
	default:
		return "Invalid value"
	}
}

func (h *User) GetUserProfile(c *gin.Context) {
	username := c.Param("username")

	resp, err := h.userService.GetUserProfile(username)
	if err != nil {
		c.JSON(404, gin.H{
			"error": "user not found",
		})
		return
	}

	c.JSON(200, resp)
}
