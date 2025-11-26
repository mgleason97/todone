package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/BurntSushi/toml"

	"todone/internal"
	"todone/internal/aggregate"
)

func main() {
	configPath := flag.String("config", "todone.toml", "Path to TODOne configuration TOML file.")
	flag.Parse()

	cfg, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("failed to load config from %s: %v", *configPath, err)
	}

	result, err := aggregate.Aggregate(cfg)
	if err != nil {
		log.Fatalf("failed to aggregate TODOs: %v", err)
	}

	fmt.Printf("TODOne aggregated %d TODOs across %d repos (messaging app=%s)\n",
		len(result.TODOs), len(cfg.Repos), cfg.Messaging.App)

	for _, t := range result.TODOs {
		fmt.Printf("\t- %s", t.Title)
	}

	fmt.Println()
}

func loadConfig(path string) (internal.Config, error) {
	var cfg internal.Config
	_, err := toml.DecodeFile(path, &cfg)
	return cfg, err
}
