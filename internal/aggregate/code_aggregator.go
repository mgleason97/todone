package aggregate

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"todone/internal"
	"todone/internal/client"
	"todone/internal/prompt"

	"github.com/openai/openai-go/packages/param"
	"github.com/openai/openai-go/responses"
)

var schema = responses.ResponseTextConfigParam{
	Format: responses.ResponseFormatTextConfigUnionParam{
		OfJSONSchema: &responses.ResponseFormatTextJSONSchemaConfigParam{
			Name:        "todo",
			Description: param.NewOpt("Structured TODO output"),
			Strict:      param.NewOpt(true),
			Schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"title":         map[string]any{"type": "string"},
					"description":   map[string]any{"type": "string"},
					"effortMinutes": map[string]any{"type": "number"},
					"priority":      map[string]any{"type": "number"},
				},
				"required":             []string{"title", "description", "effortMinutes", "priority"},
				"additionalProperties": false,
			},
		},
	},
}

// codeAggregator implements the SourceAggregator interface to extract TODOs from codebases
type codeAggregator struct {
	oai    *client.OpenAIClient
	cfg    internal.Config
	prompt string
}

func newCodeAggregator(oai *client.OpenAIClient, cfg internal.Config) *codeAggregator {
	prompt := prompt.EnrichPromptWithTask("internal/prompt/code_enrichment.md")
	return &codeAggregator{
		oai:    oai,
		cfg:    cfg,
		prompt: prompt,
	}
}

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

func (agg *codeAggregator) Extract() (RawResult, error) {
	var codeTODOs []CodeTODO
	for _, repo := range agg.cfg.Repos {
		repoCode, err := findRepoTODOs(repo)
		if err != nil {
			return nil, err
		}
		codeTODOs = append(codeTODOs, repoCode...)
	}
	return codeRawResult{TODOs: codeTODOs}, nil
}

func (agg *codeAggregator) Enrich(raw RawResult) ([]internal.TODO, error) {
	codeRaw, ok := raw.(codeRawResult)
	if !ok {
		return nil, fmt.Errorf("unexpected raw result type for code aggregator: %T", raw)
	}

	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		uniform  []internal.TODO
		errGroup error
	)

	// spawn goroutine for each enrichment
	for _, t := range codeRaw.TODOs {
		wg.Add(1)
		task := t
		go func() {
			defer wg.Done()
			enriched, err := agg.enrichCodeTask(task)
			if err != nil {
				log.Printf("Unable to enrich code task: %v", err)
				mu.Lock()
				errGroup = errors.Join(errGroup, err)
				mu.Unlock()
				return
			}
			mu.Lock()
			uniform = append(uniform, enriched)
			mu.Unlock()
		}()
	}

	wg.Wait()
	return uniform, errGroup
}

// enrichCodeTask uses an llm to make an internal.TODO from a CodeTODO
func (agg *codeAggregator) enrichCodeTask(task CodeTODO) (internal.TODO, error) {
	prompt := agg.prompt
	msg := renderCodeTask(task)

	req := client.GetResponseInput{
		SystemPrompt: prompt,
		History: []responses.ResponseInputItemUnionParam{
			client.UserMessage(msg),
		},
		ResponseFormat: schema,
	}

	res, err := agg.oai.GetResponse(context.Background(), &req)
	if err != nil {
		return internal.TODO{}, err
	}

	var enriched internal.TODO
	if err := json.Unmarshal([]byte(res.Answer), &enriched); err != nil {
		return internal.TODO{}, fmt.Errorf("failed to parse enrichment response: %w", err)
	}
	return enriched, nil
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

func renderCodeTask(task CodeTODO) string {
	var sb strings.Builder
	sb.WriteString("Code TODO found:\n")
	sb.WriteString(fmt.Sprintf("Repo: %s\nFile: %s\nLine: %d\nMatch: %s\n", task.RepoName, task.File, task.LineNumber, task.TodoLine))
	sb.WriteString("Context:\n")
	for _, line := range task.ContextLines {
		sb.WriteString(line)
		sb.WriteRune('\n')
	}
	return sb.String()
}
