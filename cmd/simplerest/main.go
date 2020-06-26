package main

import (
	"flag"
	"github.com/bramberry/simple-rest/internal/simplerest"
	"log"
)

var (
	configPath string
)

func init() {
	flag.StringVar(&configPath, "config-path", "./configs", "path to config file")
}

func main() {
	flag.Parse()

	if err := simplerest.LoadConfig(configPath); err != nil {
		log.Fatal(err)
	}

	if err := simplerest.Start(); err != nil {
		log.Fatal(err)
	}
}
