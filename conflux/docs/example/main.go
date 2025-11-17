// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package main

import (
	"os"

	"github.com/anthanhphan/gosdk/conflux"
	"github.com/anthanhphan/gosdk/logger"
)

var log = logger.NewLoggerWithFields(
	logger.String("prefix", "main"),
)

type Config struct {
	Server struct {
		Port int    `yaml:"port" json:"port"`
		Name string `yaml:"name" json:"name"`
	} `yaml:"server" json:"server"`
	Logger logger.Config `yaml:"logger" json:"logger"`
}

func main() {
	config, err := conflux.ParseConfig(
		conflux.GetConfigPathFromEnv(os.Getenv("ENV"), conflux.ExtensionYAML),
		&Config{},
	)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Infof("Config: %+v", config)
}
