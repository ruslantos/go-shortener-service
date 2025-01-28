package config

import (
	"errors"
	"flag"
	"os"
	"strconv"
	"strings"
)

var (
	FlagServerPort  string
	FlagShortURL    = "http://localhost:8080/"
	FlagLogLevel    = ""
	FileStoragePath = ""
)

type NetAddress struct {
	Host string
	Port int
}

func ParseFlags() {
	flag.StringVar(&FlagServerPort, "a", ":8080", "address and port to run server")
	flag.StringVar(&FlagLogLevel, "l", "info", "log level")
	flag.StringVar(&FileStoragePath, "f", "./internal/files/", "files storage path")

	addr := new(NetAddress)
	_ = flag.Value(addr)
	flag.Var(addr, "b", "Net address host:port")

	flag.Parse()

	envServerAddress, envBaseURL := getEnvAddresses()

	if envServerAddress != "" {
		FlagServerPort = envServerAddress
	}

	switch {
	case envBaseURL != "":
		FlagShortURL = envBaseURL + "/"
	case addr.Host != "" && addr.Port != 0:
		FlagShortURL = addr.String()
	default:
		FlagShortURL = "http://localhost:8080/"
	}

	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		FlagLogLevel = envLogLevel
	}

	if fileStoragePath := os.Getenv("FILE_STORAGE_PATH"); fileStoragePath != "" {
		FileStoragePath = fileStoragePath
	}
}

func (a NetAddress) String() string {
	return a.Host + ":" + strconv.Itoa(a.Port) + "/"
}

func (a *NetAddress) Set(s string) error {
	hp := strings.Split(s, ":")
	if len(hp) != 3 {
		return errors.New("need address in a form host:port")
	}
	port, err := strconv.Atoi(hp[2])
	if err != nil {
		return err
	}
	a.Host = hp[0] + ":" + hp[1]
	a.Port = port
	return nil
}

func getEnvAddresses() (serverAddress string, baseURL string) {
	return os.Getenv("SERVER_ADDRESS"), os.Getenv("BASE_URL")
}
