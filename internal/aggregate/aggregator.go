// Package aggregate provides aggregation logic for all TODO sources.
package aggregate

import (
	"todone/internal"
	"todone/internal/client"
)

// RawResult is a marker interface for source-specific raw TODO collections.
type RawResult any

// SourceAggregator defines the lifecycle for a single source: extract raw TODOs
// then enrich them into the uniform TODO shape.
type SourceAggregator interface {
	Extract() (RawResult, error)
	Enrich(raw RawResult) ([]internal.TODO, error)
}

// Aggregate gathers TODOs from all configured sources by running each
// SourceAggregator through its extract/enrich steps.
func Aggregate(oai *client.OpenAIClient, cfg internal.Config) ([]internal.TODO, error) {
	sources := []SourceAggregator{
		newCodeAggregator(oai, cfg),
	}

	var todos []internal.TODO
	for _, agg := range sources {
		raw, err := agg.Extract()
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
