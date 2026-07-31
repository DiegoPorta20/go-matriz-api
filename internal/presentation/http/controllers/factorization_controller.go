package controllers

import (
	"github.com/gofiber/fiber/v3"

	"github.com/DiegoPorta20/go-matriz-api/internal/application/factorization"
	"github.com/DiegoPorta20/go-matriz-api/internal/presentation/http/dto"
)

type FactorizationController struct {
	useCase *factorization.FactorizeMatrixUseCase
}

func NewFactorizationController(useCase *factorization.FactorizeMatrixUseCase) *FactorizationController {
	return &FactorizationController{useCase: useCase}
}

func (ctrl *FactorizationController) Factorize(c fiber.Ctx) error {
	var request dto.MatrixRequestDto
	if err := c.Bind().Body(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest,
			"Request body must be a JSON object with a numeric matrix")
	}

	if request.Matrix == nil {
		return fiber.NewError(fiber.StatusBadRequest, "Field matrix is required")
	}

	result, err := ctrl.useCase.Execute(c.Context(), request.Matrix)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(dto.NewFactorizationResponse(result))
}
