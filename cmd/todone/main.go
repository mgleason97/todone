package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
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
	listen(ctx, userIn)

	waitForAgent(ctx, agentOut, errCh)
}

func loadConfig(path string) (internal.Config, error) {
	var cfg internal.Config
	_, err := toml.DecodeFile(path, &cfg)
	return cfg, err
}

func listen(ctx context.Context, userIn chan<- string) {
	go func() {
		defer close(userIn)

		scanner := bufio.NewScanner(os.Stdin)
		fmt.Fprint(os.Stdout, "> ")
		for scanner.Scan() {
			line := scanner.Text()
			if strings.TrimSpace(strings.ToLower(line)) == "exit" {
				return
			}

			select {
			case <-ctx.Done():
				return
			case userIn <- line:
			}
		}
	}()
}

func waitForAgent(ctx context.Context, agentOut <-chan string, errCh <-chan error) {
	for {
		select {
		case <-ctx.Done():
			return

		case msg, ok := <-agentOut:
			if !ok {
				continue
			}
			fmt.Fprintf(os.Stdout, "agent> %s\n", msg)
			fmt.Fprint(os.Stdout, "> ")

		case err := <-errCh:
			if err != nil {
				log.Fatalf("agent failed: %v", err)
			}
			return
		}
	}
}
