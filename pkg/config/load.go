package config

import (
	"fmt"

	"github.com/spf13/viper"
)

func Load() {
	viper.SetConfigName("sidecar-config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/mnt/")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
      fmt.Printf("sidecar: Config file was not found.\n")
		} else {
			panic(fmt.Errorf("Fatal error reading config file: error: %w\n", err))
		}
	}
}
