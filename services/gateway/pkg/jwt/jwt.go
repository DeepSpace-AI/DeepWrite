package jwt

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/deepwrite/serivces/gateway/pkg/cache"
	"github.com/deepwrite/serivces/gateway/pkg/config"
	"github.com/gin-gonic/gin"
	gojwt "github.com/golang-jwt/jwt/v5"
)

const revokedTokenPrefix = "jwt:revoked:"

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	Type   string `json:"type"`
	Scope  string `json:"scope,omitempty"`
	gojwt.RegisteredClaims
}

func GenerateAccessToken(userID, email, role string) (string, error) {
	return generateToken(userID, email, role, "access", config.GetGlobalConfig().JWT.ExpireHours, "")
}

func GenerateRefreshToken(userID, email, role string) (string, error) {
	return generateToken(userID, email, role, "refresh", config.GetGlobalConfig().JWT.RefreshHours, "")
}

func GenerateAdminAccessToken(userID, email, role string) (string, error) {
	return generateToken(userID, email, role, "access", config.GetGlobalConfig().JWT.ExpireHours, "admin")
}

func GenerateAdminRefreshToken(userID, email, role string) (string, error) {
	return generateToken(userID, email, role, "refresh", config.GetGlobalConfig().JWT.RefreshHours, "admin")
}

func generateToken(userID, email, role, tokenType string, expireHours int, scope string) (string, error) {
	cfg := config.GetGlobalConfig()
	now := time.Now()
	tokenID, err := generateTokenID()
	if err != nil {
		return "", err
	}

	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		Type:   tokenType,
		Scope:  scope,
		RegisteredClaims: gojwt.RegisteredClaims{
			ID:        tokenID,
			Issuer:    cfg.JWT.Issuer,
			Subject:   userID,
			IssuedAt:  gojwt.NewNumericDate(now),
			ExpiresAt: gojwt.NewNumericDate(now.Add(time.Duration(expireHours) * time.Hour)),
		},
	}

	token := gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWT.SecretKey))
}

func ParseToken(tokenString string) (*Claims, error) {
	return parseTokenByTypeAndScope(tokenString, "access", "")
}

func ParseRefreshToken(tokenString string) (*Claims, error) {
	return parseTokenByTypeAndScope(tokenString, "refresh", "")
}

func ParseAdminToken(tokenString string) (*Claims, error) {
	return parseTokenByTypeAndScope(tokenString, "access", "admin")
}

func ParseAdminRefreshToken(tokenString string) (*Claims, error) {
	return parseTokenByTypeAndScope(tokenString, "refresh", "admin")
}

func parseTokenByTypeAndScope(tokenString, tokenType, scope string) (*Claims, error) {
	claims, err := parseSignedToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.Type != tokenType {
		return nil, errors.New("invalid token type")
	}
	if strings.TrimSpace(scope) != "" && !strings.EqualFold(strings.TrimSpace(claims.Scope), scope) {
		return nil, errors.New("invalid token scope")
	}

	revoked, err := IsTokenRevoked(tokenString)
	if err != nil {
		return nil, err
	}
	if revoked {
		return nil, errors.New("token revoked")
	}

	return claims, nil
}

func RevokeToken(tokenString string, claims *Claims) error {
	cacheStore := cache.GetDefault()
	if cacheStore == nil {
		return nil
	}

	if strings.TrimSpace(tokenString) == "" {
		return nil
	}

	if claims == nil {
		parsedClaims, err := parseSignedToken(tokenString)
		if err != nil {
			return err
		}
		claims = parsedClaims
	}

	expireTime := time.Now().Add(24 * time.Hour)
	if claims.ExpiresAt != nil {
		expireTime = claims.ExpiresAt.Time
	}

	return cacheStore.Set(revokedTokenKey(tokenString), true, expireTime)
}

func IsTokenRevoked(tokenString string) (bool, error) {
	cacheStore := cache.GetDefault()
	if cacheStore == nil {
		return false, nil
	}

	return cacheStore.Exists(revokedTokenKey(tokenString))
}

func ExtractTokenFromHeader(authHeader string) (string, error) {
	cfg := config.GetGlobalConfig()
	parts := strings.SplitN(strings.TrimSpace(authHeader), " ", 2)
	if len(parts) != 2 {
		return "", errors.New("invalid authorization header")
	}

	if !strings.EqualFold(parts[0], cfg.JWT.TokenType) {
		return "", errors.New("invalid authorization type")
	}

	if strings.TrimSpace(parts[1]) == "" {
		return "", errors.New("empty authorization token")
	}

	return parts[1], nil
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := config.GetGlobalConfig()
		headerValue := c.GetHeader(cfg.JWT.HeaderName)
		if strings.TrimSpace(headerValue) == "" {
			c.AbortWithStatusJSON(401, gin.H{"code": 401, "message": "missing authorization header"})
			return
		}

		tokenString, err := ExtractTokenFromHeader(headerValue)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"code": 401, "message": err.Error()})
			return
		}

		claims, err := ParseToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"code": 401, "message": "invalid or expired token"})
			return
		}

		c.Set("access_token_raw", tokenString)
		c.Set(cfg.JWT.ContextUserKey, claims)
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		c.Next()
	}
}

func GetCurrentClaims(c *gin.Context) (*Claims, bool) {
	cfg := config.GetGlobalConfig()
	v, ok := c.Get(cfg.JWT.ContextUserKey)
	if !ok {
		return nil, false
	}

	claims, ok := v.(*Claims)
	return claims, ok
}

func parseSignedToken(tokenString string) (*Claims, error) {
	cfg := config.GetGlobalConfig()
	if strings.TrimSpace(cfg.JWT.SecretKey) == "" {
		return nil, errors.New("jwt secret key is empty")
	}

	claims := &Claims{}
	token, err := gojwt.ParseWithClaims(tokenString, claims, func(token *gojwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*gojwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(cfg.JWT.SecretKey), nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func revokedTokenKey(tokenString string) string {
	h := sha256.Sum256([]byte(tokenString))
	return revokedTokenPrefix + hex.EncodeToString(h[:])
}

func generateTokenID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate token id failed: %w", err)
	}
	return hex.EncodeToString(b), nil
}
