package main

import (
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"github.com/go-chi/chi/v5"
)

//Переменные используем в качестве БД
var db map[string]string

var BASE_URL = `https://localhost`

//generateShort генерирует строку, которая будет использоваться длясокращения URL
func generateShort() string{
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
    "abcdefghijklmnopqrstuvwxyz" +
    "0123456789")
	length := 8
	var b strings.Builder
	for i := 0; i < length; i++ {
    	b.WriteRune(chars[rand.Intn(len(chars))])
	}
	return b.String()
}

func main() {


	db = map[string]string{}
	/*mux := http.NewServeMux()
	mux.HandleFunc(`/`, shortURL)
	mux.HandleFunc(`/{id}`, fullURL)*/

	parseFlags()

	r := chi.NewRouter()
	
	r.Get("/{id}", fullURL)
	r.Post("/", shortURL)

	err := http.ListenAndServe(flagRunAddr, r)
	if err != nil{
		panic(err)
	}
}

//shortURl обрабатывает запрос для генерации короткого URL
func shortURL(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	reqBody, err := io.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}

	_, err = url.ParseRequestURI(string(reqBody))
	if err != nil{
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte{})
	}

	res.Header().Set("Content-type", "text/plain")

	res.WriteHeader(http.StatusCreated)

	//генерируем строку для URL
	short := generateShort()

	//записываем в "БД"
	db[short] = string(reqBody)
	resBody := BASE_URL + flagBaseAddr + `/` + short
	res.Write([]byte(resBody))
}

func fullURL(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	//Получаем путь из запроса
	path := req.URL.Path
	
	short, _ := strings.CutPrefix(string(path), "/")
	short, _ = strings.CutSuffix(short, "/")

	full, ok := db[short]

	if !ok {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	res.Header().Set("Location", full)
	res.WriteHeader(http.StatusTemporaryRedirect)
	res.Write([]byte{})
}

