// Package apierr is the single source of the JSON error envelope so every
// layer (middleware, handlers) emits the identical {"error":{code,message}} shape.
package apierr

import "github.com/gofiber/fiber/v2"

// Respond writes an error envelope with the given status and code.
func Respond(c *fiber.Ctx, status int, code, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"error": fiber.Map{"code": code, "message": message},
	})
}

// Internal writes a generic 500 error envelope.
func Internal(c *fiber.Ctx) error {
	return Respond(c, fiber.StatusInternalServerError, "INTERNAL", "internal server error")
}
