package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var (
	FlagServerPort string
	FlagShortURL   = "http://localhost:8080/"
)

type NetAddress struct {
	Host string
	Port int
}

func ParseFlags() {
	flag.StringVar(&FlagServerPort, "a", ":8080", "address and port to run server")

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
		FlagShortURL = envBaseURL
	case addr.Host != "" && addr.Port != 0:
		FlagShortURL = addr.String()
	default:
		FlagShortURL = "http://localhost:8080/"
	}
	fmt.Printf("Server address: %s\n", FlagServerPort)
	fmt.Printf("BaseURL: %s\n", FlagShortURL)
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
