package cfg

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	PORT       int
	REDIS_ADDR string
	REDIS_PORT int
	HOST       string
	PORT_PG    int
	USER       string
	DB_NAME    string
	PG_PASS    string
}

func NewConfig() *Config {
	godotenv.Load()

	port := parseInt(getEnv("PORT"))
	reddisAddr := getEnv("REDIS_ADDR")
	reddisPort := parseInt(getEnv("REDIS_PORT"))
	host := getEnv("PG_HOST")
	portPG := parseInt(getEnv("PG_PORT"))
	user := getEnv("PG_USER")
	dbname := getEnv("PG_DB")
	pgPass := getEnv("PG_PASS")

	return &Config{
		PORT:       port,
		REDIS_ADDR: reddisAddr,
		REDIS_PORT: reddisPort,
		HOST:       host,
		PORT_PG:    portPG,
		USER:       user,
		DB_NAME:    dbname,
		PG_PASS:    pgPass,
	}
}

func getEnv(key string) string {
	return os.Getenv(key)
}

func parseInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
