package middleware

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func CacheControl(maxAge int) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if err := c.Next(); err != nil {
			return err
		}
		if c.Response().StatusCode() == fiber.StatusOK && c.Method() == fiber.MethodGet {
			ct := string(c.Response().Header.ContentType())
			if strings.HasPrefix(ct, "application/json") {
				c.Response().Header.Set("Cache-Control", "public, max-age="+strconv.Itoa(maxAge))
			}
		}
		return nil
	}
}
