package controllers

import (
	"github.com/gofiber/fiber/v3"

	"github.com/DiegoPorta20/go-matriz-api/internal/application/auth"
	"github.com/DiegoPorta20/go-matriz-api/internal/presentation/http/dto"
)

type AuthController struct {
	useCase *auth.IssueAccessTokenUseCase
}

func NewAuthController(useCase *auth.IssueAccessTokenUseCase) *AuthController {
	return &AuthController{useCase: useCase}
}

func (ctrl *AuthController) Login(c fiber.Ctx) error {
	var request dto.LoginRequestDto
	if err := c.Bind().Body(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest,
			"Request body must be a JSON object with username and password")
	}

	accessToken, err := ctrl.useCase.Execute(auth.Credentials{
		Username: request.Username,
		Password: request.Password,
	})
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(dto.NewAccessTokenResponse(accessToken))
}
