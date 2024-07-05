package main


import (
	"flag"
)


var flagRunAddr, flagBaseAddr string


func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", ":8080", "port to run server")
	flag.StringVar(&flagBaseAddr, "b", ":8080", "base port for short url")
	flag.Parse()
}
