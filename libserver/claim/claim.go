package claim

import (
	"os"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
)

// GetJWTClaimsHeader - Get JWT claims from header
func GetJWTClaimsHeader(c echo.Context) jwt.MapClaims {
	header := c.Request().Header.Get("Authorization")
	if header != "" {
		auth := strings.Split(header, "Bearer ")
		tokenValue := auth[1]
		token, err := jwt.Parse(tokenValue, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("jwt_secret")), nil
		})
		if err == nil && token.Valid {
			return token.Claims.(jwt.MapClaims)
		}
	}
	return nil
}

// GetJWTClaims - Get JWT claims from middleware
func GetJWTClaims(c echo.Context) jwt.MapClaims {
	test := c.Get("user")
	if test == nil {
		return GetJWTClaimsHeader(c)
	}

	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	return claims
}

// GetJWTUserID -
func GetJWTUserID(c echo.Context) int64 {
	claims := GetJWTClaims(c)
	if claims != nil {
		if len(claims) > 0 {
			v, ok := claims["user_id"].(float64)
			if ok {
				return int64(v)
			}
		}
	}
	return 0
}

// GetJWTAccountID -
func GetJWTAccountID(c echo.Context) int64 {
	claims := GetJWTClaims(c)
	if claims != nil {
		if len(claims) > 0 {
			v, ok := claims["account_id"].(float64)
			if ok {
				return int64(v)
			}
		}
	}
	return 0
}

// GetJWTAccessValue -
func GetJWTAccessValue(c echo.Context) (bool, string) {
	claims := GetJWTClaims(c)
	if claims != nil {
		if len(claims) > 0 {
			v, ok := claims["access"].(string)
			if ok {
				return true, v
			}
		}
	}
	return false, ""
}
