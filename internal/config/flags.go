package config

import (
	"errors"
	"flag"
	"os"
	"strconv"
	"strings"

	"go.uber.org/zap"

	"github.com/ruslantos/go-shortener-service/internal/middleware/logger"
)

// FlagServerPort holds the address and port to run the server.
var FlagServerPort string

// FlagShortURL holds the base URL for shortening service.
var FlagShortURL = "http://localhost:8080/"

// FlagLogLevel holds the log level for the application.
var FlagLogLevel = ""

// FileStoragePath holds the path to the file storage.
var FileStoragePath = ""

// DatabaseDsn holds the database connection string.
var DatabaseDsn = "user=videos password=password dbname=shortenerdatabase sslmode=disable"

// IsDatabaseExist indicates whether the database configuration is provided.
var IsDatabaseExist = true

// IsFileExist indicates whether the file storage configuration is provided.
var IsFileExist = true

// EnableHTTPS
var EnableHTTPS = false

// NetAddress represents a network address with a host and port.
type NetAddress struct {
	Host string
	Port int
}

// ParseFlags парсит командные строки и переменные окружения для настройки приложения.
func ParseFlags() {
	flag.StringVar(&FlagServerPort, "a", ":8080", "address and port to run server")
	flag.StringVar(&FlagLogLevel, "l", "debug", "log level")
	flag.StringVar(&FileStoragePath, "f", "", "files storage path")
	flag.StringVar(&DatabaseDsn, "d", "", "database dsn")
	flag.BoolVar(&EnableHTTPS, "s", false, "enable https")

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

	//FileStoragePath = "./tmp/links"
	// проверка конфигурации файла
	switch {
	case FileStoragePath != "":
	case os.Getenv("FILE_STORAGE_PATH") != "":
		FileStoragePath = os.Getenv("FILE_STORAGE_PATH")
	default:
		IsFileExist = false
	}

	//os.Setenv("DATABASE_DSN", "user=videos password=password dbname=shortenerdatabase sslmode=disable")
	//os.Setenv("ENABLE_HTTPS", "true")

	// проверка конфигурации БД
	switch {
	case DatabaseDsn != "":
	case os.Getenv("DATABASE_DSN") != "":
		DatabaseDsn = os.Getenv("DATABASE_DSN")
	default:
		IsDatabaseExist = false
	}

	// проверка конфигурации HTTPS
	switch {
	case os.Getenv("ENABLE_HTTPS") != "" || EnableHTTPS:
		EnableHTTPS = true
	default:
		EnableHTTPS = false
	}

	logger.GetLogger().Info("Init service config",
		zap.String("SERVER_PORT", FlagServerPort),
		zap.String("SHORT_URL", FlagShortURL),
		zap.String("LOG_LEVEL", FlagLogLevel),
		zap.String("STORAGE_PATH", FileStoragePath),
		zap.String("DATABASE_DSN", DatabaseDsn),
		zap.Boolp("IsDatabaseExist", &IsDatabaseExist),
		zap.Boolp("IsFileExist", &IsFileExist),
		zap.Boolp("EnableHTTPS", &EnableHTTPS),
	)
}

// String возвращает строковое представление сетевого адреса.
func (a NetAddress) String() string {
	return a.Host + ":" + strconv.Itoa(a.Port) + "/"
}

// Set устанавливает значение сетевого адреса из строки.
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

// getEnvAddresses возвращает адрес сервера и базовый URL из переменных окружения.
func getEnvAddresses() (serverAddress string, baseURL string) {
	return os.Getenv("SERVER_ADDRESS"), os.Getenv("BASE_URL")
}
