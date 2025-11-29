package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/BurntSushi/toml"

	"todone/internal"
	"todone/internal/agent"
	"todone/internal/client"
)

func main() {
	configPath := flag.String("config", "todone.toml", "Path to TODOne configuration TOML file.")
	flag.Parse()

	cfg, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("failed to load config from %s: %v", *configPath, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	userIn := make(chan string)
	agentOut := make(chan string)
	ag, err := agent.New(client.NewOpenAIClient(), userIn, agentOut, cfg)
	if err != nil {
		log.Fatalf("failed to init agent: %v", err)
	}

	errCh := make(chan error, 1)
	go func() {
		defer close(agentOut)
		errCh <- ag.Run(ctx)
	}()

	go func() {
		defer close(userIn)

		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := scanner.Text()

			select {
			case <-ctx.Done():
				return
			case userIn <- line:
			}
		}

	}()

	for {
		select {
		case <-ctx.Done():
			return

		case msg, ok := <-agentOut:
			if !ok {
				// Agent finished; wait for error (if any)
				continue
			}

			fmt.Fprintf(os.Stdout, "agent> %s\n", msg)

		case err := <-errCh:
			if err != nil {
				log.Fatalf("agent failed: %v", err)
			}
			return
		}
	}
}

func loadConfig(path string) (internal.Config, error) {
	var cfg internal.Config
	_, err := toml.DecodeFile(path, &cfg)
	return cfg, err
}
