package api

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/sirupsen/logrus"
)

type MyCustomClaims struct {
	jwt.StandardClaims
	ClientId string
}

func issueJwt(clientId string, hmacSampleSecret []byte) (string, error) {
	// logrus.Info(string(hmacSampleSecret))
	// Create a new token object, specifying signing method and the claims
	// you would like it to contain.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, MyCustomClaims{
		ClientId: clientId,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(72 * time.Hour).Unix(),
		},
		//"nbf": time.Now().Add(72 * time.Hour).Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret
	return token.SignedString(hmacSampleSecret)
}

func authJwt(tokenStr string, secret []byte) (*jwt.Token, error) {

	// Remove bearer things
	realToken := strings.Replace(tokenStr, "Bearer ", "", 1)
	// logrus.Info(string(secret))

	// We need to ensure to handle clock drifts properly
	timeOffset := "60"
	timeOffsetEnv := os.Getenv("JWT_SUPPORTED_TIME_OFFSET_MINS")
	if timeOffsetEnv != "" {
		timeOffset = timeOffsetEnv
	}

	if timeOffset != "" {
		realTimeOffset, err := strconv.Atoi(timeOffset)
		if err != nil {
			logrus.Warn("Failed to pase JWT_SUPPORTED_TIME_OFFSET_MINS=", timeOffsetEnv)
		} else {
			// logrus.Info("Adjusting time offset to ", realTimeOffset, " minutes")
			jwt.TimeFunc = func() time.Time {
				return time.Now().Add(-time.Minute * time.Duration(realTimeOffset))
			}
		}

	}

	// Parse takes the token string and a function for looking up the key. The latter is especially
	// useful if you use multiple keys for your application.  The standard is to use 'kid' in the
	// head of the token to identify which key to use, but the parsed token (head and claims) is provided
	// to the callback, providing flexibility.
	token, err := jwt.Parse(realToken, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return secret, nil
	})

	if err != nil {
		return nil, err
	}

	if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return token, nil
	} else {
		return nil, fmt.Errorf("Jwt: Token invalid")
	}
}
