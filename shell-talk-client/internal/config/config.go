package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		URL string `mapstructure:"url"`
	} `mapstructure:"server"`
}

var Cfg *Config

func LoadConfig() {
	viper.AddConfigPath("./configs") // 설정 파일 경로
	viper.SetConfigName("config")    // 설정 파일 이름
	viper.SetConfigType("yaml")      // 설정 파일 타입

	viper.AutomaticEnv() // 환경 변수도 읽기

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %s", err)
	}

	if err := viper.Unmarshal(&Cfg); err != nil {
		log.Fatalf("Unable to decode config into struct: %v", err)
	}
}
