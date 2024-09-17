package config

import (
	"errors"
	"log"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Jwt       JWT
	Cassandra CassandraDB
	Log       Logger
}

type JWT struct {
	Secret         string
	Issuer         string
	Audience       []string
	ExpireInHours  time.Duration
	LeewayInSecond time.Duration
}

type CassandraDB struct {
	UserName string
	Password string
	KeySpace string
	Hosts    []string
}

type Logger struct {
	FilePath string
	Level    int
}

// the path from the parent folder exp1 as an example
func GetConfig(path string) (*Config, error) {
	v, err := getViper(path)
	if err != nil {
		return nil, err
	}
	c, err := parseViperConfig(v)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func getViper(path string) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigName(path)
	v.AddConfigPath("../")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, errors.New("config file not found")
		} else {
			return nil, err
		}
	}
	return v, nil
}

func parseViperConfig(v *viper.Viper) (*Config, error) {
	var cfg Config
	err := v.Unmarshal(&cfg)
	if err != nil {
		log.Printf("Unable to parse config: %v", err)
		return nil, err
	}
	return &cfg, nil
}
