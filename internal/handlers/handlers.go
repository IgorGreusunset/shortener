package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/IgorGreusunset/shortener/cmd/config"
	model "github.com/IgorGreusunset/shortener/internal/app"
	"github.com/IgorGreusunset/shortener/internal/helpers"
	"github.com/IgorGreusunset/shortener/internal/logger"
	"github.com/IgorGreusunset/shortener/internal/storage"
)

// Handler для обработки Post-запроса на запись новой URL структуры в хранилище
func PostHandler(db storage.Repository) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		reqBody, err := io.ReadAll(req.Body)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		defer req.Body.Close()

		logger.Log.Debugln(string(reqBody))

		//Проверяем, что в теле запроса корректный URL-адрес
		_, err = url.ParseRequestURI(string(reqBody))
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		//Генерируем ID для короткой ссылки
		id := helpers.Generate()

		//Создаем новый экземпляр URL структуры и записываем его в хранилище
		urlToAdd := model.NewURL(id, string(reqBody))
		userID, _ := req.Cookie("userID")
		logger.Log.Debugln(userID)
		urlToAdd.UserID = userID.Value
		ctx := context.Background()
		if err := db.Create(ctx, urlToAdd); err != nil {
			var uee *storage.URLExistsError
			if errors.As(err, &uee) {
				res.Header().Set("Content-type", "text/plain")
				res.WriteHeader(http.StatusConflict)
				resBody := config.Base + `/` + uee.ShortURL
				if _, err := res.Write([]byte(resBody)); err != nil {
					log.Printf("Error writing response: %v\n", err)
					http.Error(res, "Internal server error", http.StatusInternalServerError)
				}
				return
			}
			logger.Log.Debugln(err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		//Записываем заголовок и тело ответа
		res.Header().Set("Content-type", "text/plain")
		res.WriteHeader(http.StatusCreated)
		resBody := config.Base + `/` + id
		if _, err := res.Write([]byte(resBody)); err != nil {
			log.Printf("Error writing response: %v\n", err)
			http.Error(res, "Internal server error", http.StatusInternalServerError)
		}
	}
}

// Handler для обработки Get-запроса на получение ссылки по ID
func GetByIDHandler(db storage.Repository) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		//Получаем ID из запроса и ищем по нему URL структуру в хранилище
		short := chi.URLParam(req, "id")

		fullURL, ok := db.GetByID(short)

		if !ok {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		//Записываем заголовок ответа
		res.Header().Set("Location", fullURL.FullURL)
		res.WriteHeader(http.StatusTemporaryRedirect)
	}
}

// Handler для обработки json-запроса на создание новой ссылки
func APIPostHandler(db storage.Repository) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		//Получаем данные для создания URL модели из запроса
		var urlFromRequest model.APIPostRequest
		dec := json.NewDecoder(req.Body)
		if err := dec.Decode(&urlFromRequest); err != nil {
			logger.Log.Debugln("error", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		//Проверяем корректость адреса в теле запроса
		_, err := url.ParseRequestURI(urlFromRequest.URL)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		id := helpers.Generate()

		//Создаем модель и записываем в storage
		urlToAdd := model.NewURL(id, urlFromRequest.URL)
		userID, _ := req.Cookie("userID")
		urlToAdd.UserID = userID.Value
		ctx := context.Background()
		if err := db.Create(ctx, urlToAdd); err != nil {
			var uee *storage.URLExistsError
			if errors.As(err, &uee) {
				res.Header().Set("Content-type", "application/json")
				res.WriteHeader(http.StatusConflict)
				result := config.Base + `/` + uee.ShortURL
				resp := model.NewAPIPostResponse(result)
				response, err := json.Marshal(resp)
				if err != nil {
					logger.Log.Debugln("error", err)
					res.WriteHeader(http.StatusInternalServerError)
					return
				}
				res.Write(response)
				return
			}
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		//Формируем и сериализируем тело ответа
		result := config.Base + `/` + id
		resp := model.NewAPIPostResponse(result)
		response, err := json.Marshal(resp)
		if err != nil {
			logger.Log.Debugln("error", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		//Записываем заголовок и тело ответа
		res.Header().Set("Content-type", "application/json")
		res.WriteHeader(http.StatusCreated)
		res.Write(response)
	}
}

// Handler для проверки подключения к БД
func PingHandler(db storage.Repository) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		err := db.Ping()
		if err == nil {
			res.WriteHeader(http.StatusOK)
		} else {
			res.WriteHeader(http.StatusInternalServerError)
		}
	}
}

// Handler для добавления списка ссылок
func BathcHandler(db storage.Repository) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		var (
			requests []model.APIBatchRequest
			urls     []model.URL
			shorts   []model.APIBatchResponse
		)

		//Десериализуем тело запроса в слайс
		err := json.NewDecoder(req.Body).Decode(&requests)
		if err != nil {
			http.Error(res, "Failed decoding request body", http.StatusBadRequest)
		}

		userID, _ := req.Cookie("userID")

		//Проходим по слайсу и для каждого элемента создаем model.URL и подготавливаем модель для ответа
		for _, r := range requests {
			sh := helpers.Generate()
			url := model.NewURL(sh, r.URL)
			url.UserID = userID.Value
			urls = append(urls, *url)
			w := model.NewAPIBatchResponse(r.ID, config.Base+`/`+sh)
			shorts = append(shorts, *w)
		}

		ctx := context.Background()
		//Сохраняем ссылки в хранилище
		if len(urls) != 0 {
			if err = db.CreateBatch(ctx, urls); err != nil {
				http.Error(res, "Failed to save urls in db", http.StatusInternalServerError)
			}
		}
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusCreated)

		//Сериализируем тело ответа
		if err = json.NewEncoder(res).Encode(shorts); err != nil {
			http.Error(res, "Error during encoding response", http.StatusInternalServerError)
		}
	}

}

func URLByUserHandler(db storage.Repository) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		var resBody []model.UsersURLsResponse
		userID, err := req.Cookie("userID")
		if errors.Is(err, http.ErrNoCookie) {
			res.WriteHeader(http.StatusUnauthorized)
			return
		} else if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		urls, err := db.UsersURLs(userID.Value)
		if err != nil {
			logger.Log.Errorln("error during db request to get all users urls", err)
			res.WriteHeader(http.StatusInternalServerError)
		}

		if len(urls) == 0 {
			res.WriteHeader(http.StatusNoContent)
			return
		}

		for _, u := range urls {
			r := model.NewUsersURLsResponse(config.Base+`/`+u.ID, u.FullURL)
			resBody = append(resBody, *r)
		}

		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)
		if err = json.NewEncoder(res).Encode(resBody); err != nil {
			http.Error(res, "Error during encoding response", http.StatusInternalServerError)
		}
	}
}
