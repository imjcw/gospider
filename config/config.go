package config

import "github.com/spf13/viper"

func init() {
	viper.SetConfigFile("./config.yml")
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
}
