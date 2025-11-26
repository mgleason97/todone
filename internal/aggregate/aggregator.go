// Package aggregate provides aggregation logic for all TODO sources
package aggregate

import "todone/internal"

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

// TODO: aggregation should just find raw TODO sources and their context
// E.g. just the lines around a TODO commment.
// After aggregation, use an llm to enrich each TODO into a formal type with
// title, desc, effort, pri
func enrich() {}
