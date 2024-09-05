package middleware

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

const TokenExp = time.Hour * 24
const SecretKey = "dmTMd2N9"

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

func WithAuth(h http.Handler) http.Handler {
	authFN := func(w http.ResponseWriter, r *http.Request) {

		t, err := r.Cookie("Auth")
		if errors.Is(err, http.ErrNoCookie) {
			generateSetCookies(w, r)
			h.ServeHTTP(w, r)
			return
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			h.ServeHTTP(w, r)
			return
		}

		id := GetUserID(t.Value)
		if id == "" {
			generateSetCookies(w, r)
			h.ServeHTTP(w, r)
			return
		} else {
			http.SetCookie(w, &http.Cookie{
				Name:    "userID",
				Value:   id,
				Path:    "/",
				Expires: time.Now().Add(TokenExp),
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
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		return "", "", err
	}
	return tokenString, userID, nil
}

// Проверяем токен и получаем пользователя
func GetUserID(tokenString string) string {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})
	if err != nil {
		return ""
	}

	if !token.Valid {
		return ""
	}
	return claims.UserID
}

func generateSetCookies(w http.ResponseWriter, req *http.Request) {
	token, id, err := buildJWTString()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "Auth",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(TokenExp),
		HttpOnly: true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:    "userID",
		Value:   id,
		Path:    "/",
		Expires: time.Now().Add(TokenExp),
	})
	req.Header.Set("Authorization", token)
}

func WithUserID(h http.Handler) http.Handler {
	userFN := func(res http.ResponseWriter, req *http.Request) {
		userID, err := req.Cookie("userID")
		if errors.Is(err, http.ErrNoCookie) {
			http.Error(res, "Unauthorized", http.StatusUnauthorized)
			return
		} else if err != nil {
			http.Error(res, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(req.Context(), "UserID", userID.Value)
		h.ServeHTTP(res, req.WithContext(ctx))

	}

	return http.HandlerFunc(userFN)
}