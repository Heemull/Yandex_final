package auth

import (
	"fmt"
	"net/http"
	"os"

	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("super-secret-key-change-in-production")

// Claims структура полезной нагрузки JWT
type Claims struct {
	PasswordHash string `json:"ph"`
	jwt.RegisteredClaims
}

// GenerateToken создает JWT токен
func GenerateToken(password string) (string, error) {
	envPass := os.Getenv("TODO_PASSWORD")
	if envPass == "" {
		return "", fmt.Errorf("authentication is disabled")
	}

	if password != envPass {
		return "", fmt.Errorf("invalid password")
	}

	claims := &Claims{
		PasswordHash: envPass,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(8 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ValidateToken проверяет валидность токена
func ValidateToken(tokenString string) bool {
	envPass := os.Getenv("TODO_PASSWORD")
	if envPass == "" {
		return true
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return false
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return false
	}

	if claims.PasswordHash != envPass {
		return false
	}

	return true
}

// Middleware для Chi
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		envPass := os.Getenv("TODO_PASSWORD")

		// Если пароль не установлен, пропускаем запрос
		if envPass == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Получаем куки
		cookie, err := r.Cookie("token")
		if err != nil {
			// Возвращаем JSON ошибку
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"Authentication required"}`))
			return
		}

		if !ValidateToken(cookie.Value) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"Invalid or expired token"}`))
			return
		}

		next.ServeHTTP(w, r)
	})
}
