package main

import (
	"log"
	"net/http"
	"os"

	"github.com/IgorGreusunset/shortener/cmd/config"
	model "github.com/IgorGreusunset/shortener/internal/app"
	"github.com/IgorGreusunset/shortener/internal/handlers"
	"github.com/IgorGreusunset/shortener/internal/logger"
	"github.com/IgorGreusunset/shortener/internal/middleware"
	"github.com/IgorGreusunset/shortener/internal/storage"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)


func main() {

	config.ParseFlag()

	router := chi.NewRouter()

	logger.Initialize()

	var db storage.Repository

	switch config.DataBase {
		case "" :
			file, err := os.OpenFile(config.File, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
			if err != nil {
				log.Fatalf("Error during opening file with shorten urls: %v", err)
			}
			database := storage.NewStorage(map[string]model.URL{})
			database.SetFile(file)
	
				//Наполняем хранилище данными из файла
			err = database.FillFromFile(file)
			if err != nil {
				logger.Log.Infof("Error during reading from file with shorten urls: %v", err)
			}
			file.Close()

			db = database

		default:
			var err error
			database, err := storage.NewDatabase(config.DataBase)
			if err != nil {
				log.Fatalf("Error during database connection: %v", err)
			}
			defer database.DB.Close()

			db = database

	}
	

	//Обертки для handlers, чтобы использовать их в роутере
	PostHandlerWrapper := func (res http.ResponseWriter, req *http.Request)  {
		handlers.PostHandler(db, res, req)
	}

	GetHandlerWrapper := func (res http.ResponseWriter, req *http.Request)  {
		handlers.GetByIDHandler(db, res, req)
	}

	APIPostHandlerWrapper := func (res http.ResponseWriter, req *http.Request)  {
		handlers.APIPostHandler(db, res, req)
	}

	PingHandlerWrapper := func(res http.ResponseWriter, req *http.Request) {
		handlers.PingHandler(db, res, req)
	}

	BatchHandlerWrapper := func(res http.ResponseWriter, req *http.Request) {
		handlers.BathcHandler(db, res, req)
	}

	//Подключаем middlewares
	router.Use(middleware.WithLogging)
	router.Use(middleware.GzipMiddleware)

	router.Post(`/`, PostHandlerWrapper)
	router.Get(`/{id}`, GetHandlerWrapper)
	router.Post(`/api/shorten`, APIPostHandlerWrapper)
	router.Get(`/ping`, PingHandlerWrapper)
	router.Post(`/api/shorten/batch`, BatchHandlerWrapper)

	serverAdd := config.Serv

	log.Fatal(http.ListenAndServe(serverAdd, router))
}


