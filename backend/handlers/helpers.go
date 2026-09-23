package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	DB *pgxpool.Pool
}

func HashPassword(pw string) string {
	sum := sha256.Sum256([]byte("cbt-smk:" + pw))
	return hex.EncodeToString(sum[:])
}

func RandomToken(n int) string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, n)
	for i := range b {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		b[i] = chars[idx.Int64()]
	}
	return string(b)
}

func ParseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(strings.TrimSpace(s))
}

func Bad(c *fiber.Ctx, msg string) error {
	return c.Status(400).JSON(fiber.Map{"error": msg})
}

func Fail(c *fiber.Ctx, err error) error {
	return c.Status(500).JSON(fiber.Map{"error": err.Error()})
}

func Shuffle[T any](in []T) []T {
	out := make([]T, len(in))
	copy(out, in)
	for i := len(out) - 1; i > 0; i-- {
		jBig, _ := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		j := int(jBig.Int64())
		out[i], out[j] = out[j], out[i]
	}
	return out
}
