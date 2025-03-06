package conf

import (
	"flag"
	"os"
	"strconv"
	"time"
)

const defaultSecretKeyTime = 35

type Config struct {
	BndAdr        string
	AcrBndAdr     string
	AcrPoint      string
	DSN           string
	SecretKey     string
	SecretKeyTime time.Duration
}

func (s *Config) ParseFlags() {
	flag.StringVar(&s.BndAdr, "a", "localhost:8080", "host where server is run")
	flag.StringVar(&s.DSN, "d", "", "database dsn")
	flag.StringVar(&s.SecretKey, "k", "DEFAULT_SECRET_KEY", "Secret key for JWT")
	flag.StringVar(&s.AcrBndAdr, "aa", "http://localhost:8090", "Host where accural service run.")
	flag.StringVar(&s.AcrPoint, "ap", "/api/orders", "Accurual point")
	flag.DurationVar(&s.SecretKeyTime, "kt", defaultSecretKeyTime*time.Minute, "Time secret key in minutes")
}

func (s *Config) ParseEnv() {
	if env := os.Getenv("RUN_ADDRESS"); env != "" {
		s.BndAdr = env
	}

	if env := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); env != "" {
		s.AcrBndAdr = env
	}

	if env := os.Getenv("ACCURAL_POINT"); env != "" {
		s.AcrPoint = env
	}

	if env := os.Getenv("DATABASE_URI"); env != "" {
		s.DSN = env
	}

	if env := os.Getenv("SECRET_KEY"); env != "" {
		s.SecretKey = env
	}

	if env := os.Getenv("SECRET_KEY_TIME"); env != "" {
		duration, err := strconv.Atoi(env)

		if err == nil {
			s.SecretKeyTime = time.Duration(duration) * time.Minute
		} else {
			s.SecretKeyTime = time.Duration(defaultSecretKeyTime) * time.Minute
		}
	}
}

func NewConf() Config {
	c := Config{}
	c.ParseFlags()
	flag.Parse()
	c.ParseEnv()
	return c
}
