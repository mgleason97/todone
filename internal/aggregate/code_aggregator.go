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

type codeAggregator struct{}

// CodeTODO represents a TODO found in code with surrounding context.
type CodeTODO struct {
	RepoName     string
	File         string
	LineNumber   int
	TodoLine     string
	ContextLines []string
}

type codeRawResult struct {
	TODOs []CodeTODO
}

func (codeAggregator) Extract(cfg internal.Config) (RawResult, error) {
	raw, err := findRawCodeTODOs(cfg.Repos)
	if err != nil {
		return nil, err
	}
	return codeRawResult{TODOs: raw}, nil
}

func (codeAggregator) Enrich(raw RawResult) ([]internal.TODO, error) {
	codeRaw, ok := raw.(codeRawResult)
	if !ok {
		return nil, fmt.Errorf("unexpected raw result type for code aggregator: %T", raw)
	}
	return enrichCodeTODOs(codeRaw.TODOs)
}

// findRawCodeTODOs returns raw code TODOs (with context) from all repos.
func findRawCodeTODOs(repos []internal.Repo) ([]CodeTODO, error) {
	var codeTODOs []CodeTODO
	for _, repo := range repos {
		repoCode, err := findRepoTODOs(repo)
		if err != nil {
			return nil, err
		}
		codeTODOs = append(codeTODOs, repoCode...)
	}
	return codeTODOs, nil
}

// enrichCodeTODOs converts raw code TODOs into uniform TODOs.
// This is where an LLM-based enrichment pass can add title/description/effort/priority.
func enrichCodeTODOs(raw []CodeTODO) ([]internal.TODO, error) {
	todos := make([]internal.TODO, 0, len(raw))
	for _, item := range raw {
		todos = append(todos, internal.TODO{
			Title:         item.TodoLine,
			Description:   fmt.Sprintf("Found in %s:%s:%d", item.RepoName, item.File, item.LineNumber),
			EffortMinutes: 0,
			Priority:      1,
		})
	}
	return todos, nil
}

func findRepoTODOs(repo internal.Repo) ([]CodeTODO, error) {
	cmd := exec.Command("rg", "--line-number", "--no-heading", "--context", "5", "TODO", repo.Path)
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

	blocks := splitRipgrepBlocks(stdout.String())
	var codeTODOs []CodeTODO

	for _, block := range blocks {
		items, err := parseRipgrepBlock(repo.Name, block)
		if err != nil {
			return nil, err
		}
		codeTODOs = append(codeTODOs, items...)
	}

	return codeTODOs, nil
}

func splitRipgrepBlocks(output string) [][]string {
	var blocks [][]string
	var current []string
	for _, line := range strings.Split(strings.TrimSuffix(output, "\n"), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if line == "--" {
			if len(current) > 0 {
				blocks = append(blocks, current)
				current = nil
			}
			continue
		}
		current = append(current, line)
	}
	if len(current) > 0 {
		blocks = append(blocks, current)
	}
	return blocks
}

func parseRipgrepBlock(repoName string, block []string) ([]CodeTODO, error) {
	if len(block) == 0 {
		return nil, nil
	}

	var contextLines []string
	for _, line := range block {
		_, _, text, _, err := splitLine(line)
		if err != nil {
			return nil, err
		}
		contextLines = append(contextLines, text)
	}

	var results []CodeTODO
	for _, line := range block {
		file, lineNum, text, isMatch, err := splitLine(line)
		if err != nil {
			return nil, err
		}
		if !isMatch {
			continue
		}
		results = append(results, CodeTODO{
			RepoName:     repoName,
			File:         file,
			LineNumber:   lineNum,
			TodoLine:     text,
			ContextLines: contextLines,
		})
	}

	return results, nil
}

func splitLine(line string) (file string, lineNum int, text string, isMatch bool, err error) {
	tryParts := func(sep string) (f string, ln int, t string, ok bool, e error) {
		parts := strings.SplitN(line, sep, 3)
		if len(parts) < 3 {
			return "", 0, "", false, nil
		}
		ln, e = strconv.Atoi(parts[1])
		if e != nil {
			return "", 0, "", false, nil
		}
		return parts[0], ln, strings.TrimSpace(parts[2]), true, nil
	}

	if f, ln, t, ok, e := tryParts(":"); ok {
		return f, ln, t, true, e
	}
	if f, ln, t, ok, e := tryParts("-"); ok {
		return f, ln, t, false, e
	}

	return "", 0, "", false, errors.New("unexpected ripgrep output: " + line)
}
