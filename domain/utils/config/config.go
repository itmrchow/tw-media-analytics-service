package config

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// InitConfig 初始化 config.
func InitConfig() {
	viper.AutomaticEnv()
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal().Err(err).Msg("config init error")
	}

	log.Info().Msgf("config init success")
}
