package config

import (
	"encoding/json"
	"errors"
	"flag"
	"os"
	"strconv"
	"strings"

	"go.uber.org/zap"

	"github.com/ruslantos/go-shortener-service/internal/middleware/logger"
)

var FlagShortURL = "http://localhost:8080/"

// Config содержит все параметры конфигурации приложения
type Config struct {
	ServerPort      string
	BaseURL         string
	LogLevel        string
	FileStoragePath string
	DatabaseDsn     string
	IsDatabaseExist bool
	IsFileExist     bool
	EnableHTTPS     bool
	ConfigFile      string
}

type ConfigFile struct {
	ServerAddress   string `json:"server_address"`    // -a / SERVER_ADDRESS
	BaseURL         string `json:"base_url"`          // -b / BASE_URL
	FileStoragePath string `json:"file_storage_path"` // -f / FILE_STORAGE_PATH
	DatabaseDSN     string `json:"database_dsn"`      // -d / DATABASE_DSN
	EnableHTTPS     bool   `json:"enable_https"`      // -s / ENABLE_HTTPS
}

// NetAddress represents a network address with a host and port.
type NetAddress struct {
	Host string
	Port int
}

// ParseFlags парсит командные строки и переменные окружения для настройки приложения.
func ParseFlags() Config {
	c := Config{
		ServerPort:      ":8080",
		BaseURL:         "",
		LogLevel:        "",
		DatabaseDsn:     "",
		IsDatabaseExist: true,
		IsFileExist:     true,
		EnableHTTPS:     false,
	}

	flag.StringVar(&c.ServerPort, "a", ":8080", "address and port to run server")
	flag.StringVar(&c.LogLevel, "l", "debug", "log level")
	flag.StringVar(&c.FileStoragePath, "f", "", "files storage path")
	flag.StringVar(&c.DatabaseDsn, "d", "", "database dsn")
	flag.BoolVar(&c.EnableHTTPS, "s", false, "enable https")
	flag.StringVar(&c.ConfigFile, "c", "", "config file")

	addr := new(NetAddress)
	_ = flag.Value(addr)
	flag.Var(addr, "b", "Net address host:port")

	flag.Parse()

	configFile, err := c.loadConfigFromFile()
	if err != nil {
		logger.GetLogger().Error("Failed to load config from file", zap.Error(err))
	}

	// server address
	switch {
	case c.ServerPort != "":
	case os.Getenv("SERVER_ADDRESS") != "":
		c.ServerPort = os.Getenv("SERVER_ADDRESS")
	case configFile.ServerAddress != "":
		c.ServerPort = configFile.ServerAddress
	default:
		c.ServerPort = ":8080"
	}

	// base URL
	switch {
	case c.BaseURL != "":
		c.BaseURL = c.BaseURL + "/"
	case addr.Host != "" && addr.Port != 0:
		c.BaseURL = addr.String()
	//case addr.Port == 0:
	case os.Getenv("BASE_URL") != "":
		c.BaseURL = os.Getenv("BASE_URL")
	case configFile.BaseURL != "":
		c.BaseURL = configFile.BaseURL
	default:
		c.BaseURL = "http://localhost:8080/"
	}
	FlagShortURL = c.BaseURL

	// log level
	switch {
	case c.LogLevel != "":
	case os.Getenv("LOG_LEVEL") != "":
		c.LogLevel = os.Getenv("LOG_LEVEL")
	default:
		c.LogLevel = "debug"
	}

	// file storage path
	switch {
	case c.FileStoragePath != "":
	case os.Getenv("FILE_STORAGE_PATH") != "":
		c.FileStoragePath = os.Getenv("FILE_STORAGE_PATH")
	case configFile.FileStoragePath != "":
		c.FileStoragePath = configFile.FileStoragePath
	default:
		c.IsFileExist = false
	}

	//os.Setenv("DATABASE_DSN", "user=videos password=password dbname=shortenerdatabase sslmode=disable")

	// database dsn
	switch {
	case c.DatabaseDsn != "":
	case os.Getenv("DATABASE_DSN") != "":
		c.DatabaseDsn = os.Getenv("DATABASE_DSN")
	case configFile.DatabaseDSN != "":
		c.DatabaseDsn = configFile.DatabaseDSN
	default:
		c.IsDatabaseExist = false
	}

	// enable HTTPS
	switch {
	case c.EnableHTTPS:
	case os.Getenv("ENABLE_HTTPS") != "":
		c.EnableHTTPS = getBoolEnv("ENABLE_HTTPS", false)
	case configFile.EnableHTTPS:
		c.EnableHTTPS = configFile.EnableHTTPS
	default:
		c.EnableHTTPS = false
	}

	logger.GetLogger().Info("Init service config",
		zap.String("SERVER_PORT", c.ServerPort),
		zap.String("BASE_URL", c.BaseURL),
		zap.String("LOG_LEVEL", c.LogLevel),
		zap.String("STORAGE_PATH", c.FileStoragePath),
		zap.String("DATABASE_DSN", c.DatabaseDsn),
		zap.Boolp("IsDatabaseExist", &c.IsDatabaseExist),
		zap.Boolp("IsFileExist", &c.IsFileExist),
		zap.Boolp("EnableHTTPS", &c.EnableHTTPS),
	)

	return c
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

func (c *Config) loadConfigFromFile() (ConfigFile, error) {
	var config ConfigFile
	filePath := c.getConfigFilePath()
	if filePath == "" {
		return config, nil
	}

	file, err := os.ReadFile(filePath)
	if err != nil {
		return config, err
	}

	if err := json.Unmarshal(file, &config); err != nil {
		return config, err
	}

	return config, nil
}

func (c *Config) getConfigFilePath() string {
	switch {
	case c.ConfigFile != "":
		return c.ConfigFile
	case os.Getenv("CONFIG") != "":
		return os.Getenv("CONFIG")
	default:
		return ""
	}
}

func getBoolEnv(key string, defaultVal bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return strings.ToLower(val) == "true" || val == "1"
}
