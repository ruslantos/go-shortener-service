package config

import (
	"flag"
)

var (
	FlagRunAddr  string
	FlagShortUrl string
)

func ParseFlags() {
	flag.StringVar(&FlagRunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&FlagShortUrl, "b", "http://localhost:8080/", "short url address")
	flag.Parse()
}
