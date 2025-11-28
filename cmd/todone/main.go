package main

import (
	"context"
	"flag"
	"log"
	"strings"
	"time"

	"github.com/BurntSushi/toml"

	"todone/internal"
	"todone/internal/agent"
	"todone/internal/client"
)

func main() {
	configPath := flag.String("config", "todone.toml", "Path to TODOne configuration TOML file.")
	question := flag.String("question", "", "Optional question to ask the agent about your TODOs.")
	flag.Parse()

	cfg, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("failed to load config from %s: %v", *configPath, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if strings.TrimSpace(*question) == "" {
		return
	}

	userPipe := make(chan string)
	agent, err := agent.New(client.NewOpenAIClient(), userPipe, cfg)
	if err != nil {
		log.Fatalf("failed to init agent: %v", err)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- agent.Run(ctx)
	}()

	userPipe <- *question
	close(userPipe)

	if err := <-errCh; err != nil {
		log.Fatalf("agent failed: %v", err)
	}
}

func loadConfig(path string) (internal.Config, error) {
	var cfg internal.Config
	_, err := toml.DecodeFile(path, &cfg)
	return cfg, err
}
