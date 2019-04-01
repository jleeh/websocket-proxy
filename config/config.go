package config

import (
	"github.com/spf13/viper"
)

// Config holds the configuration values for the proxy server
type Config struct {
	Port           int      `mapstructure:"port"`
	Server         string   `mapstructure:"server"`
	AuthType       string   `mapstructure:"auth_type"`
	KeyManagerType string   `mapstructure:"key_manager_type"`
	KeyIdentifier  string   `mapstructure:"key_identifier"`
	AllowedOrigins []string `mapstructure:"allowed_origins"`
}

// New creates a new configuration instance via viper from a file and/or env vars
func New(filename string, defaults map[string]interface{}) *Config {
	v := viper.New()
	c := Config{}
	for key, value := range defaults {
		v.SetDefault(key, value)
	}
	v.SetConfigName(filename)
	v.AddConfigPath(".")

	v.AutomaticEnv()
	_ = v.ReadInConfig()
	_ = v.Unmarshal(&c)
	return &c
}
