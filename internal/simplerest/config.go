package simplerest

import (
	"fmt"
	"github.com/spf13/viper"
)

// Config is global object that holds all application level variables.
var Config appConfig

// Config ...
type appConfig struct {
	BindPort              string `mapstructure:"bindPort"`
	LogLevel              string `mapstructure:"logLevel"`
	DatabaseURL           string `mapstructure:"databaseURL"`
	SessionKey            string `mapstructure:"sessionKey"`
	AWSBucketName         string `mapstructure:"awsBucketName"`
	AWSPicturesFolderName string `mapstructure:"awsPicturesFolderName"`
}

// LoadConfig ...
func LoadConfig(configPaths ...string) error {
	v := viper.New()
	v.SetConfigName("simplerest")
	v.SetConfigType("yaml")
	v.AutomaticEnv()
	for _, path := range configPaths {
		v.AddConfigPath(path)
	}
	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read the configuration file: %s", err)
	}
	return v.Unmarshal(&Config)
}
