package aggregate

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"todone/internal"
)

// Result groups all aggregated TODOs from any source.
type Result struct {
	TODOs []internal.TODO
}

// Aggregate gathers TODOs from all configured sources.
func Aggregate(cfg internal.Config) (Result, error) {
	codeItems, err := aggregateCodeTODOs(cfg.Repos)
	if err != nil {
		return Result{}, err
	}

	return Result{
		TODOs: codeItems,
	}, nil
}

func aggregateCodeTODOs(repos []internal.Repo) ([]internal.TODO, error) {
	var todos []internal.TODO
	for _, repo := range repos {
		repoItems, err := findRepoTODOs(repo)
		if err != nil {
			return nil, err
		}
		todos = append(todos, repoItems...)
	}
	return todos, nil
}

func findRepoTODOs(repo internal.Repo) ([]internal.TODO, error) {
	cmd := exec.Command("rg", "--line-number", "--no-heading", "TODO", repo.Path)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			// ripgrep uses exit code 1 when no matches are found; treat as empty result.
			return nil, nil
		}
		return nil, err
	}

	var items []internal.TODO
	for _, line := range strings.Split(strings.TrimSuffix(stdout.String(), "\n"), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		item, err := parseRipgrepLine(repo.Name, line)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func parseRipgrepLine(repoName, line string) (internal.TODO, error) {
	parts := strings.SplitN(line, ":", 3)
	if len(parts) < 3 {
		return internal.TODO{}, errors.New("unexpected ripgrep output: " + line)
	}

	lineNum, err := strconv.Atoi(parts[1])
	if err != nil {
		return internal.TODO{}, err
	}

	text := strings.TrimSpace(parts[2])
	return internal.TODO{
		Title:         text,
		Description:   fmt.Sprintf("Found in %s:%s:%d", repoName, parts[0], lineNum),
		EffortMinutes: 0,
		Priority:      1,
	}, nil
}
