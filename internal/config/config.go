package config

import (
	"os"
	"strconv"
	"time"
)

// Config 服务配置
type Config struct {
	Port         int    `json:"port"`
	Host         string `json:"host"`
	MaxRooms     int    `json:"max_rooms"`
	ReadTimeout  int    `json:"read_timeout"`  // 秒
	WriteTimeout int    `json:"write_timeout"` // 秒

	AppBaseURL string     `json:"app_base_url"` // 用于生成邮件验证链接
	DB         DBConfig   `json:"db"`
	JWT        JWTConfig  `json:"jwt"`
	SMTP       SMTPConfig `json:"smtp"`
}

// DBConfig 数据库配置
type DBConfig struct {
	DSN string `json:"dsn"` // MYSQL_DSN
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret string        `json:"-"`   // JWT_SECRET，不序列化
	TTL    time.Duration `json:"ttl"` // JWT_TTL_HOURS 默认 24h
}

// SMTPConfig 邮件发送配置
type SMTPConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"-"`
	From     string `json:"from"`
	DevMode  bool   `json:"dev_mode"` // SMTP_DEV_MODE=true 时改为日志打印
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	smtpUser := getEnvString("SMTP_USERNAME", "")
	cfg := &Config{
		Port:         getEnvInt("SERVER_PORT", 8080),
		Host:         getEnvString("SERVER_HOST", "0.0.0.0"),
		MaxRooms:     getEnvInt("MAX_ROOMS", 100),
		ReadTimeout:  getEnvInt("READ_TIMEOUT", 60),
		WriteTimeout: getEnvInt("WRITE_TIMEOUT", 60),
		AppBaseURL:   getEnvString("APP_BASE_URL", "http://localhost:8080"),
		DB: DBConfig{
			DSN: getEnvString("MYSQL_DSN", ""),
		},
		JWT: JWTConfig{
			Secret: getEnvString("JWT_SECRET", ""),
			TTL:    time.Duration(getEnvInt("JWT_TTL_HOURS", 24)) * time.Hour,
		},
		SMTP: SMTPConfig{
			Host:     getEnvString("SMTP_HOST", ""),
			Port:     getEnvInt("SMTP_PORT", 587),
			Username: smtpUser,
			Password: getEnvString("SMTP_PASSWORD", ""),
			From:     getEnvString("SMTP_FROM", smtpUser),
			DevMode:  getEnvBool("SMTP_DEV_MODE", false),
		},
	}
	return cfg
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

func getEnvBool(key string, defaultVal bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	b, err := strconv.ParseBool(val)
	if err != nil {
		return defaultVal
	}
	return b
}
