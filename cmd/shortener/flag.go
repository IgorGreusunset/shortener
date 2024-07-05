package main


import (
	"flag"
)


var flagRunAddr, flagBaseAddr string


func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", "http://localhost:8080", "port to run server")
	flag.StringVar(&flagBaseAddr, "b", "http://localhost:8080", "base port for short url")
	flag.Parse()
	if flagRunAddr != ":8080" {
		flagBaseAddr = flagRunAddr
	} else if flagBaseAddr != ":8080" {
		flagRunAddr = flagBaseAddr
	}
}
