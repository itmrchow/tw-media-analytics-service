package infra

import "github.com/rs/zerolog"

var (
	logger *zerolog.Logger
)

func SetInfraLogger(zLogger *zerolog.Logger) {
	logger = zLogger
}
