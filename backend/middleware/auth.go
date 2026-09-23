package middleware

import (
	"os"
	"strings"

	"cbt-api/models"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func secret() []byte {
	s := os.Getenv("JWT_SECRET")
	if s == "" {
		s = "dev-secret-change-me"
	}
	return []byte(s)
}

func Sign(claims models.Claims) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":    claims.UserID,
		"student_id": claims.StudentID,
		"teacher_id": claims.TeacherID,
		"exam_id":    claims.ExamID,
		"role":       claims.Role,
		"full_name":  claims.FullName,
	})
	return t.SignedString(secret())
}

func Parse(tokenStr string) (*models.Claims, error) {
	t, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return secret(), nil
	})
	if err != nil || !t.Valid {
		return nil, fiber.ErrUnauthorized
	}
	mc, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fiber.ErrUnauthorized
	}
	str := func(k string) string {
		v, _ := mc[k].(string)
		return v
	}
	return &models.Claims{
		UserID:    str("user_id"),
		StudentID: str("student_id"),
		TeacherID: str("teacher_id"),
		ExamID:    str("exam_id"),
		Role:      str("role"),
		FullName:  str("full_name"),
	}, nil
}

func Protected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		h := c.Get("Authorization")
		if h == "" {
			return c.Status(401).JSON(fiber.Map{"error": "token diperlukan"})
		}
		raw := strings.TrimPrefix(h, "Bearer ")
		claims, err := Parse(raw)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "token tidak valid"})
		}
		c.Locals("claims", claims)
		return c.Next()
	}
}

func Role(roles ...string) fiber.Handler {
	allowed := map[string]bool{}
	for _, r := range roles {
		allowed[r] = true
	}
	return func(c *fiber.Ctx) error {
		claims, _ := c.Locals("claims").(*models.Claims)
		if claims == nil || !allowed[claims.Role] {
			return c.Status(403).JSON(fiber.Map{"error": "akses ditolak"})
		}
		return c.Next()
	}
}

func GetClaims(c *fiber.Ctx) *models.Claims {
	v, _ := c.Locals("claims").(*models.Claims)
	return v
}
