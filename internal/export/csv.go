// Package export provides CSV export for Neo4j bulk import
package export

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/zachbornheimer/morpheme-extraction/internal/graph"
)

// ExportToCSV exports semantic graph to CSV files for Neo4j bulk import
// Creates two files: nodes.csv (words) and relationships.csv (edges)
func ExportToCSV(g *graph.SemanticGraph, nodesPath, relsPath string) error {
	// Export nodes
	if err := exportNodes(g, nodesPath); err != nil {
		return fmt.Errorf("failed to export nodes: %w", err)
	}

	// Export relationships
	if err := exportRelationships(g, relsPath); err != nil {
		return fmt.Errorf("failed to export relationships: %w", err)
	}

	return nil
}

func exportNodes(g *graph.SemanticGraph, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"wordId:ID", "text", "centerFreq:int", ":LABEL"}); err != nil {
		return err
	}

	// Write nodes
	nodes := g.GetAllNodes()
	for _, node := range nodes {
		if node.Word() == nil {
			continue
		}

		record := []string{
			fmt.Sprintf("%d", node.Word().ID),
			node.Word().Str,
			fmt.Sprintf("%d", node.Word().Frequency.GetCenter()),
			"Word",
		}

		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

func exportRelationships(g *graph.SemanticGraph, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{":START_ID", ":END_ID", "weight:float", ":TYPE"}); err != nil {
		return err
	}

	// Write relationships (only once per edge)
	nodes := g.GetAllNodes()
	for _, node := range nodes {
		if node.Word() == nil {
			continue
		}

		neighbors := node.GetNeighborsWithWeights()
		for neighborID, weight := range neighbors {
			// Only write each edge once (avoid duplicates)
			if node.Word().ID < neighborID {
				record := []string{
					fmt.Sprintf("%d", node.Word().ID),
					fmt.Sprintf("%d", neighborID),
					fmt.Sprintf("%.1f", weight),
					"RELATED",
				}

				if err := writer.Write(record); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
