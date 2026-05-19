package api

import (
	"os"
	"strings"
)

type Environment string

const (
	Test        Environment = "test"
	Development Environment = "development"
	Production  Environment = "production"
)

func Env() Environment {
	if strings.HasSuffix(os.Args[0], ".test") || os.Args[1] == "-test.run" {
		return Test
	}
	v := os.Getenv("APP_ENV")
	if v == "" {
		return Development
	}
	return Environment(v)
}

func IsEnv(e Environment) bool {
	return Env() == e
}
