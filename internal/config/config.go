package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config 全局配置
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
}

type ServerConfig struct {
	Addr    string
	GinMode string
}

type DatabaseConfig struct {
	Enabled      bool
	Host         string
	Port         int
	User         string
	Password     string
	Name         string
	MaxOpenConns int
	MaxIdleConns int
}

// DSN 生成 MySQL DSN
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		d.User, d.Password, d.Host, d.Port, d.Name)
}

// Load 加载配置,优先环境变量,fallback 到默认值
func Load() *Config {
	// .env 可选,失败不退出
	_ = godotenv.Load()

	return &Config{
		Server: ServerConfig{
			Addr:    getEnv("HTTP_ADDR", ":8182"),
			GinMode: getEnv("GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Enabled:      getEnvBool("DB_ENABLED", true),
			Host:         getEnv("DB_HOST", "127.0.0.1"),
			Port:         getEnvInt("DB_PORT", 3306),
			User:         getEnv("DB_USER", "root"),
			Password:     getEnv("DB_PASSWORD", ""),
			Name:         getEnv("DB_NAME", "hexis"),
			MaxOpenConns: 20,
			MaxIdleConns: 10,
		},
	}
}

func getEnvBool(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err == nil {
			return parsed
		}
		if value == "1" {
			return true
		}
		if value == "0" {
			return false
		}
	}
	return fallback
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
