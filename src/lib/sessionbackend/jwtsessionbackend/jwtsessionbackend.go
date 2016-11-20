package jwtsessionbackend

import (
	"fmt"
	"github.com/clawio/clawiod/src/lib/sessionbackend"
	"github.com/clawio/clawiod/src/proto"
	"github.com/dgrijalva/jwt-go"
	"gopkg.in/ini.v1"
	"time"
)

const backendID = "jwt"

type backend struct {
	config *ini.File
}

func New(config *ini.File) sessionbackend.SessionBackend {
	return &backend{config}
}

func (u *backend) GetBackendID() string {
	return backendID
}

func (u *backend) GenerateSessionTicket(user *proto.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username":     user.Username,
		"display_name": user.DisplayName,
		"exp":          time.Now().Unix() + 3600, // 1 hour
		"nbf":          time.Date(2015, 10, 10, 12, 0, 0, 0, time.UTC).Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(u.config.Section("").Key("sessionbackend.jwt.secret").MustString("")))
	return tokenString, err
}

func (u *backend) DecodeSessionTicket(ticket string) (*proto.User, error) {
	// Parse takes the token string and a function for looking up the key. The latter is especially
	// useful if you use multiple keys for your application.  The standard is to use 'kid' in the
	// head of the token to identify which key to use, but the parsed token (head and claims) is provided
	// to the callback, providing flexibility.
	token, err := jwt.Parse(ticket, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(u.config.Section("").Key("sessionbackend.jwt.secret").MustString("")), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		u := &proto.User{
			Username:    claims["username"].(string),
			DisplayName: claims["display_name"].(string),
		}
		return u, nil
	}
	return nil, err

}
