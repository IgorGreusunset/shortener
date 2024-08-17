package middleware

import (
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"net/http"
	"time"
)

const TOKEN_EXP = time.Hour * 24
const SECRET_KEY = "dmTMd2N9"

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

func WithAuth(h http.Handler) http.Handler {
	authFN := func(w http.ResponseWriter, r *http.Request) {

		t, err := r.Cookie("Auth")
		if errors.Is(err, http.ErrNoCookie) {
			generateSetCookies(w)
			h.ServeHTTP(w, r)
			return
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			h.ServeHTTP(w, r)
			return
		}

		id := getUserID(t.Value)
		if id == "" {
			generateSetCookies(w)
			h.ServeHTTP(w, r)
			return
		} else {
			http.SetCookie(w, &http.Cookie{
				Name:    "userID",
				Value:   id,
				Path:    "/",
				Expires: time.Now().Add(TOKEN_EXP),
			})
		}

		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(authFN)
}

func buildJWTString() (string, string, error) {
	userID := uuid.New().String()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TOKEN_EXP)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", "", err
	}
	return tokenString, userID, nil
}

func getUserID(tokenString string) string {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(SECRET_KEY), nil
	})
	if err != nil {
		return ""
	}

	if !token.Valid {
		return ""
	}
	return claims.UserID
}

func generateSetCookies(w http.ResponseWriter) {
	token, id, err := buildJWTString()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "Auth",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(TOKEN_EXP),
		HttpOnly: true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:    "userID",
		Value:   id,
		Path:    "/",
		Expires: time.Now().Add(TOKEN_EXP),
	})
}
