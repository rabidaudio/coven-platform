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

func looksLikeTest() bool {
	return strings.HasSuffix(os.Args[0], ".test") || (len(os.Args) > 1 && os.Args[1] == "-test.run")
}

func Env() Environment {
	if looksLikeTest() {
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
