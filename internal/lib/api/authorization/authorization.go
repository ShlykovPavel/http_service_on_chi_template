package authorization

import (
	"github.com/ShlykovPavel/http_service_on_chi_template/internal/lib/jwt_tokens"
	"github.com/golang-jwt/jwt/v5"
)

// Authorization проверяет предоставленный токен и получает аргументы тела токена
func Authorization(tokenString string, secretKey string) (jwt.MapClaims, error) {
	jwtClaims, err := jwt_tokens.VerifyToken(tokenString, secretKey)
	if err != nil {
		return nil, err
	}
	return jwtClaims, nil
}
