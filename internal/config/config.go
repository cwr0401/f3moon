package config

import (
	"os"
	"strconv"
)

// Config 服务配置
type Config struct {
	Port         int    `json:"port"`
	Host         string `json:"host"`
	MaxRooms     int    `json:"max_rooms"`
	ReadTimeout  int    `json:"read_timeout"`  // 秒
	WriteTimeout int    `json:"write_timeout"` // 秒
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		Port:         getEnvInt("SERVER_PORT", 8080),
		Host:         getEnvString("SERVER_HOST", "0.0.0.0"),
		MaxRooms:     getEnvInt("MAX_ROOMS", 100),
		ReadTimeout:  getEnvInt("READ_TIMEOUT", 60),
		WriteTimeout: getEnvInt("WRITE_TIMEOUT", 60),
	}
}

func getEnvInt(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return n
}

func getEnvString(key string, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}
