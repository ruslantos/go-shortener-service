package config

import (
	"errors"
	"flag"
	"strconv"
	"strings"
)

var (
	FlagRunAddr  string
	FlagShortURL = "http://localhost:8080/"
)

type NetAddress struct {
	Host string
	Port int
}

func ParseFlags() {
	flag.StringVar(&FlagRunAddr, "a", ":8080", "address and port to run server")

	addr := new(NetAddress)
	_ = flag.Value(addr)
	flag.Var(addr, "b", "Net address host:port")

	flag.Parse()

	if addr.Host == "" {
		return
	} else {
		FlagShortURL = addr.String()
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
	a.Host = hp[0] + hp[1]
	a.Port = port
	return nil
}
