package config

import (
	"fmt"

	"github.com/spf13/viper"
)

func LoadConfig() {
	configFilename := "config.yaml"

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.sidecar")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Printf("Config file %s was not found.\n", configFilename)
		} else {
			panic(fmt.Errorf("Fatal error reading config file %s: error: %w\n",
				configFilename, err))
		}
	}
}
