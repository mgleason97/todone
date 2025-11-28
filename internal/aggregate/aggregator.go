// Package aggregate provides aggregation logic for all TODO sources.
package aggregate

import "todone/internal"

// RawResult is a marker interface for source-specific raw TODO collections.
type RawResult any

// SourceAggregator defines the lifecycle for a single source: extract raw TODOs
// then enrich them into the uniform TODO shape.
type SourceAggregator interface {
	Extract(cfg internal.Config) (RawResult, error)
	Enrich(raw RawResult) ([]internal.TODO, error)
}

// Aggregate gathers TODOs from all configured sources by running each
// SourceAggregator through its extract/enrich steps.
func Aggregate(cfg internal.Config) ([]internal.TODO, error) {
	sources := []SourceAggregator{
		codeAggregator{},
	}

	var todos []internal.TODO
	for _, agg := range sources {
		raw, err := agg.Extract(cfg)
		if err != nil {
			return nil, err
		}

		enriched, err := agg.Enrich(raw)
		if err != nil {
			return nil, err
		}
		todos = append(todos, enriched...)
	}

	return todos, nil
}
