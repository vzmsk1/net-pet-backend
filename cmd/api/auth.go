package main

import (
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"net/http"
	"time"
)

type Auth struct {
	Issuer        string
	Audience      string
	Secret        string
	TokenExpiry   time.Duration
	RefreshExpiry time.Duration
	CookieDomain  string
	CookiePath    string
	CookieName    string
}

type jwtUser struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type TokenPairs struct {
	Token        string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type Claims struct {
	jwt.RegisteredClaims
}

func (j *Auth) GenerateTokenPair(user *jwtUser) (TokenPairs, error) {
	// create a token
	token := jwt.New(jwt.SigningMethodHS256)

	// set the claims
	claims := token.Claims.(jwt.MapClaims)
	claims["name"] = fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	claims["sub"] = fmt.Sprint(user.ID)
	claims["aud"] = j.Audience
	claims["iss"] = j.Issuer
	claims["iat"] = time.Now().UTC().Unix()
	claims["typ"] = "JWT"

	// set the expiry for jwt
	claims["exp"] = time.Now().UTC().Add(j.TokenExpiry).Unix()

	// create a signed token
	signedAccessToken, err := token.SignedString([]byte(j.Secret))
	if err != nil {
		return TokenPairs{}, err
	}

	// create a refresh token and set claims
	refreshToken := jwt.New(jwt.SigningMethodHS256)
	refreshTokenClaims := refreshToken.Claims.(jwt.MapClaims)
	refreshTokenClaims["sub"] = fmt.Sprint(user.ID)
	refreshTokenClaims["iat"] = time.Now().UTC().Unix()

	// set the expiry for the refresh token
	refreshTokenClaims["exp"] = time.Now().UTC().Add(j.TokenExpiry).Unix()

	// create signed refresh token
	signedRefreshToken, err := refreshToken.SignedString([]byte(j.Secret))
	if err != nil {
		return TokenPairs{}, err
	}

	// create TokenPairs and populate with signed tokens
	var tokenPairs = TokenPairs{
		Token:        signedAccessToken,
		RefreshToken: signedRefreshToken,
	}

	// return TokenPairs
	return tokenPairs, nil
}

func (j *Auth) GetRefreshCookie(refreshToken string) *http.Cookie {
	return &http.Cookie{
		Name:     j.CookieName,
		Path:     j.CookiePath,
		Value:    refreshToken,
		Expires:  time.Now().Add(j.RefreshExpiry),
		MaxAge:   int(j.TokenExpiry.Seconds()),
		SameSite: http.SameSiteStrictMode,
		Domain:   j.CookieDomain,
		HttpOnly: true,
		Secure:   true,
	}
}

func (j *Auth) GetExpiredRefreshCookie() *http.Cookie {
	return &http.Cookie{
		Name:     j.CookieName,
		Path:     j.CookiePath,
		Value:    "",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		SameSite: http.SameSiteStrictMode,
		Domain:   j.CookieDomain,
		HttpOnly: true,
		Secure:   true,
	}
}
