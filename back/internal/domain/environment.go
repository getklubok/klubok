package domain

import (
	"strings"
)

type Environment string

func (t Environment) String() string {
	return string(t)
}

const (
	EnvironmentLocal Environment = "LOCAL"
	EnvironmentStage Environment = "STAGE"
	EnvironmentProd  Environment = "PROD"
)

func EnvironmentFromString(s string) Environment {
	e := Environment(strings.ToUpper(s))

	switch e {
	case EnvironmentLocal, EnvironmentStage, EnvironmentProd:
		return e
	}

	return EnvironmentProd
}
