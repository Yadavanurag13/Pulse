// services/user-service/internal/utils/jwt/jwt.go
package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"health-tracker-project/services/user-service/internal/utils/logger" // Import the logger
)

// JWTSecret is the secret key used for signing JWTs.
// *** IMPORTANT: In a real production environment, this should be read from an environment variable! ***
// var jwtSecret = []byte(os.Getenv("JWT_SECRET")) // Use this in production
var jwtSecret = []byte("your-super-secret-jwt-key") // For development convenience, CHANGE THIS!

// Claims struct holds custom claims along with standard JWT claims.
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"` // Keeping 'Username' in claims for display/identification
	jwt.RegisteredClaims
}

// GenerateJWT generates a new JWT token for a given user.
func GenerateJWT(userID, username string, expiration time.Duration) (string, error) {
	expirationTime := time.Now().Add(expiration)
	claims := &Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		logger.Logger.Errorf("Failed to sign JWT token for user ID '%s': %v", userID, err)
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	logger.Logger.Debugf("JWT token generated for user ID: %s", userID)
	return tokenString, nil
}

// ParseJWT parses and validates a JWT token string.
func ParseJWT(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		logger.Logger.Debugf("JWT token parsing failed: %v", err)
		return nil, fmt.Errorf("token parsing failed: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		logger.Logger.Warn("Invalid JWT token claims or token not valid.")
		return nil, fmt.Errorf("invalid token claims")
	}

	logger.Logger.Debugf("JWT token parsed successfully for user ID: %s", claims.UserID)
	return claims, nil
}