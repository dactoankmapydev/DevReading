package security

import (
	"devread/model"

	"os"
	"time"

	"github.com/golang-jwt/jwt"
)

/*func CreateToken(userID string) (*model.TokenDetails, error) {
	td := &model.TokenDetails{
		AtExpires:   time.Now().Add(time.Minute * 15).Unix(),
		AccessUUID:  uuid.New().String(),
		RtExpires:   time.Now().Add(time.Hour * 24).Unix(),
		RefreshUUID: uuid.New().String(),
	}

	var err error

	// creating access token
	atClaims := jwt.MapClaims{
		"access_uuid": td.AccessUUID,
		"user_id":     userID,
		"exp":         td.AtExpires,
	}
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	td.AccessToken, err = at.SignedString([]byte(os.Getenv("ACCESS_SECRET")))
	if err != nil {
		return nil, err
	}

	// creatig refresh token
	rtClaims := jwt.MapClaims{
		"refresh_uuid": td.RefreshUUID,
		"user_id":      userID,
		"exp":          td.RtExpires,
	}
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	td.RefreshToken, err = rt.SignedString([]byte(os.Getenv("REFRESH_SECRET")))
	if err != nil {
		return nil, err
	}

	return td, nil
}*/

func CreateToken(user model.User) (string, error) {
	claims := &model.TokenDetails{
		UserID: user.UserID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(os.Getenv("ACCESS_SECRET")))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
