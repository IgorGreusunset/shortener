package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/IgorGreusunset/shortener/cmd/config"
	model "github.com/IgorGreusunset/shortener/internal/app"
	"github.com/IgorGreusunset/shortener/internal/helpers"
	"github.com/IgorGreusunset/shortener/internal/logger"
	"github.com/IgorGreusunset/shortener/internal/storage"
	"github.com/go-chi/chi/v5"
)

// Handler для обработки Post-запроса на запись новой URL структуры в хранилище
func PostHandler(db storage.Repository, res http.ResponseWriter, req *http.Request) {

	reqBody, err := io.ReadAll(req.Body)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	defer req.Body.Close()

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
	if err := db.Create(urlToAdd); err != nil {
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

// Handler для обработки Get-запроса на получение ссылки по ID
func GetByIDHandler(db storage.Repository, res http.ResponseWriter, req *http.Request) {

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

//Handler для обработки json-запроса на создание новой ссылки
func APIPostHandler(db storage.Repository, res http.ResponseWriter, req *http.Request) {

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
	if err := db.Create(urlToAdd); err != nil {
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

func PingHandler(db storage.Repository, res http.ResponseWriter, req *http.Request) {
	err := db.Ping()
	if err == nil {
		res.WriteHeader(http.StatusOK)
	} else {
		res.WriteHeader(http.StatusInternalServerError)
	}
}

func BathcHandler(db storage.Repository, res http.ResponseWriter, req *http.Request) {

	var (
		requests []model.APIBatchRequest
		urls []model.URL
		shorts []model.APIBatchResponse
	)

	err := json.NewDecoder(req.Body).Decode(&requests)
	if err != nil {
		http.Error(res, "Failed decoding request body", http.StatusBadRequest)
	}

	for _, r := range requests {
		sh := helpers.Generate()
		url := model.NewURL(sh, r.URL)
		urls = append(urls, *url)
		w := model.NewAPIBatchResponse(r.ID, config.Base + `/` + sh)
		shorts = append(shorts, *w)
	}

	if len(urls) != 0 {
		if err = db.CreateBatch(urls); err != nil {
			http.Error(res, "Failed to save urls in db", http.StatusInternalServerError)
		}
	}
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)

	if err = json.NewEncoder(res).Encode(shorts); err != nil {
		http.Error(res, "Error during encoding response", http.StatusInternalServerError)
	}

	
}
