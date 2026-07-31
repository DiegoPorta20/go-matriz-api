package controllers

import (
	"github.com/gofiber/fiber/v3"

	"github.com/DiegoPorta20/go-matriz-api/internal/presentation/http/dto"
)

type HealthController struct{}

func NewHealthController() HealthController {
	return HealthController{}
}

func (HealthController) Check(c fiber.Ctx) error {
	return c.JSON(dto.HealthResponseDto{Status: "ok"})
}
