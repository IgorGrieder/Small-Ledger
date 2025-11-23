package cfg

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	APPLICATION_PORT int
	REDIS_ADDR       string
	REDIS_PORT       int
	PG_HOST          string
	PG_PORT          int
	PG_USER          string
	PG_DB_NAME       string
	PG_PASS          string
	CURRENCY_URL     string
}

func NewConfig() *Config {
	godotenv.Load()

	port := parseInt(getEnv("APPLICATION_PORT"))
	reddisAddr := getEnv("REDIS_ADDR")
	reddisPort := parseInt(getEnv("REDIS_PORT"))
	host := getEnv("PG_HOST")
	portPG := parseInt(getEnv("PG_PORT"))
	user := getEnv("PG_USER")
	dbname := getEnv("PG_DB")
	pgPass := getEnv("PG_PASS")
	currencyUrl := getEnv("CURRENCY_URL")

	return &Config{
		APPLICATION_PORT: port,
		REDIS_ADDR:       reddisAddr,
		REDIS_PORT:       reddisPort,
		PG_HOST:          host,
		PG_PORT:          portPG,
		PG_USER:          user,
		PG_DB_NAME:       dbname,
		PG_PASS:          pgPass,
		CURRENCY_URL:     currencyUrl,
	}
}

func getEnv(key string) string {
	return os.Getenv(key)
}

func parseInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
