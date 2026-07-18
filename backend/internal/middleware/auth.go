package middleware

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	ShopID string `json:"shop_id"`
	jwt.RegisteredClaims
}

func Auth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "error": "unauthorized"})
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "error": "invalid token"})
		}

		c.Locals("userID", claims.UserID)
		c.Locals("role", claims.Role)
		c.Locals("shopID", claims.ShopID)
		return c.Next()
	}
}

func RequireRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := c.Locals("role").(string)
		for _, r := range roles {
			if r == role {
				return c.Next()
			}
		}
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": "forbidden"})
	}
}

// RequireShop rejects requests from users who do not belong to a shop.
// This prevents users without a shop_id (e.g. legacy seed admin) from accessing
// shop-scoped resources.
func RequireShop() fiber.Handler {
	return func(c *fiber.Ctx) error {
		shopID := c.Locals("shopID")
		if shopID == nil || shopID.(string) == "" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": "no shop assigned"})
		}
		return c.Next()
	}
}
