package middleware

import (
	"crypto/rsa"
	"errors"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

var (
	errMissingToken = errors.New("missing token")
	errInvalidToken = errors.New("token is expired or invalid")
)

// jwtPublicKey is the RSA public key loaded once at startup.
var jwtPublicKey *rsa.PublicKey

// LoadJWTPublicKey reads the PEM-encoded RSA public key from the path
// specified in JWT_PUBLIC_KEY_PATH. Call this once at application startup.
func LoadJWTPublicKey() error {
	path := os.Getenv("JWT_PUBLIC_KEY_PATH")
	if path == "" {
		path = "keys/public.pem"
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	key, err := jwt.ParseRSAPublicKeyFromPEM(data)
	if err != nil {
		return err
	}
	jwtPublicKey = key
	return nil
}

// JWTMiddleware validates the RS256 Bearer token and populates Fiber locals:
// "userID" (string), "plan" (string), "tz" (string).
func JWTMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			return unauthorized(c)
		}
		raw := strings.TrimPrefix(header, "Bearer ")

		parseOpts := []jwt.ParserOption{
			jwt.WithValidMethods([]string{"RS256"}),
			jwt.WithExpirationRequired(),
		}
		if iss := os.Getenv("JWT_ISSUER"); iss != "" {
			parseOpts = append(parseOpts, jwt.WithIssuer(iss))
		}
		if aud := os.Getenv("JWT_AUDIENCE"); aud != "" {
			parseOpts = append(parseOpts, jwt.WithAudience(aud))
		}

		token, err := jwt.Parse(raw, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, errInvalidToken
			}
			return jwtPublicKey, nil
		}, parseOpts...)
		if err != nil || !token.Valid {
			return unauthorized(c)
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return unauthorized(c)
		}

		userID, _ := claims["sub"].(string)
		plan, _ := claims["plan"].(string)
		tz, _ := claims["tz"].(string)

		if userID == "" {
			return unauthorized(c)
		}

		c.Locals("userID", userID)
		c.Locals("plan", plan)
		c.Locals("tz", tz)

		return c.Next()
	}
}

func unauthorized(c *fiber.Ctx) error {
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		"error":   "unauthorized",
		"message": "Token is expired or invalid",
	})
}
