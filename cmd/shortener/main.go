package main

import (
	"context"
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

	if config.DataBase == "" {
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
	} else {
		var err error
		database, err := storage.NewDatabase(config.DataBase)
		if err != nil {
			log.Fatalf("Error during database connection: %v", err)
		}
		defer database.DB.Close()

		db = database

	}

	delCh := make(chan model.DeleteTask)
	delList := make([]model.DeleteTask, 0)

	go func() {
		for task := range delCh {
				delList = append(delList, task)
				if len(delList) == 15 {
					db.Delete(context.Background(), delList)
				}
			}
		
	}()

	//Подключаем middlewares
	router.Use(middleware.WithLogging)
	router.Use(middleware.GzipMiddleware)
	router.Use(middleware.WithAuth)
	//router.Use(middleware.WithUserID)

	router.Post(`/`, handlers.PostHandler(db))
	router.Get(`/{id}`, handlers.GetByIDHandler(db))
	router.Post(`/api/shorten`, handlers.APIPostHandler(db))
	router.Get(`/ping`, handlers.PingHandler(db))
	router.Post(`/api/shorten/batch`, handlers.BathcHandler(db))
	router.Get(`/api/user/urls`, handlers.URLByUserHandler(db))
	router.Delete(`/api/user/urls`, handlers.DeleteBatchURLsHandler(db, delCh))

	serverAdd := config.Serv

	log.Fatal(http.ListenAndServe(serverAdd, router))
}
