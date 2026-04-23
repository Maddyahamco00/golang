package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrUnauthorized = errors.New("unauthorized")
)

type JWTMiddleware struct {
	secret     string
	expiration time.Duration
}

func NewJWTMiddleware(secret string) *JWTMiddleware {
	return &JWTMiddleware{
		secret:     secret,
		expiration: 24 * time.Hour, // Token valid for 24 hours
	}
}

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func (m *JWTMiddleware) GenerateToken(userID, email, role string) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secret))
}

func (m *JWTMiddleware) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(m.secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

func (m *JWTMiddleware) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		claims, err := m.ValidateToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// HMAC middleware for request signing
type HMACMiddleware struct {
	secret string
}

func NewHMACMiddleware(secret string) *HMACMiddleware {
	return &HMACMiddleware{secret: secret}
}

// GenerateSignature creates HMAC signature for a request
func (m *HMACMiddleware) GenerateSignature(method, path, body, timestamp string) string {
	message := strings.Join([]string{method, path, body, timestamp}, "|")
	h := hmac.New(sha256.New, []byte(m.secret))
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

// ValidateSignature validates HMAC signature
func (m *HMACMiddleware) ValidateSignature(method, path, body, timestamp, signature string) bool {
	expected := m.GenerateSignature(method, path, body, timestamp)
	return hmac.Equal([]byte(expected), []byte(signature))
}

// HMACAuth middleware for sensitive endpoints
func (m *HMACMiddleware) HMACAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		signature := c.GetHeader("X-Signature")
		timestamp := c.GetHeader("X-Timestamp")

		if signature == "" || timestamp == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "signature and timestamp required"})
			c.Abort()
			return
		}

		// Check timestamp is within 5 minutes
		ts, err := time.Parse(time.RFC3339, timestamp)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid timestamp"})
			c.Abort()
			return
		}

		if time.Since(ts) > 5*time.Minute {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "request expired"})
			c.Abort()
			return
		}

		// Read body
		body := ""
		if c.Request.Body != nil {
			bodyBytes := make([]byte, c.Request.ContentLength)
			c.Request.Body.Read(bodyBytes)
			body = string(bodyBytes)
			// Restore body for downstream handlers
			c.Request.Body = nil
		}

		// Validate signature
		if !m.ValidateSignature(c.Request.Method, c.Request.URL.Path, body, timestamp, signature) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimiter simple in-memory rate limiter
type RateLimiter struct {
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (rl *RateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use IP or user ID as key
		key := c.ClientIP()
		if userID, exists := c.Get("user_id"); exists {
			key = userID.(string)
		}

		now := time.Now()

		// Clean old requests
		var valid []time.Time
		for _, t := range rl.requests[key] {
			if now.Sub(t) < rl.window {
				valid = append(valid, t)
			}
		}

		if len(valid) >= rl.limit {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			c.Abort()
			return
		}

		// Add current request
		rl.requests[key] = append(valid, now)

		c.Next()
	}
}