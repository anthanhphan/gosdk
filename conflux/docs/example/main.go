// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package main

import (
	"github.com/anthanhphan/gosdk/conflux"
	"github.com/anthanhphan/gosdk/logger"
)

var log = logger.NewLoggerWithFields(
	logger.String("prefix", "main"),
)

type Config struct {
	Server struct {
		Port int    `yaml:"port" json:"port" validate:"required,min=1,max=65535"`
		Name string `yaml:"name" json:"name" validate:"required"`
	} `yaml:"server" json:"server"`
	Logger logger.Config `yaml:"logger" json:"logger"`
}

func main() {
	// MustLoad parses the config file AND validates all struct tags.
	// If any validation rule fails (e.g., port out of range), it panics.
	config := conflux.MustLoad[Config]("./config/config.local.yaml")

	log.Infof("Config: %+v", config)
}
