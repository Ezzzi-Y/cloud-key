//go:build tools

package cloudkey

import (
	_ "github.com/gin-gonic/gin"
	_ "github.com/golang-jwt/jwt/v5"
	_ "github.com/pquerna/otp"
	_ "golang.org/x/crypto/bcrypt"
	_ "gorm.io/driver/sqlite"
)
