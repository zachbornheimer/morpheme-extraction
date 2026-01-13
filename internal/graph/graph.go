// Package graph provides an explicit semantic graph structure for word relationships
// Graph represents ground truth: nodes are words, edges are co-occurrence relationships
// This is NOT a vector embedding - meaning emerges from graph structure (symbolic)
package graph

import (
	"sync"

	"github.com/zachbornheimer/morpheme-extraction/internal/dictionary"
)

// SemanticGraph represents the complete semantic network
// Nodes are words (WordIdentity), edges are similarity relationships
// Thread-safe for concurrent access during graph construction
type SemanticGraph struct {
	nodes map[int]*SemanticNode              // WordIdentity.ID → node
	words map[int]*dictionary.WordIdentity   // Fast ID → Word lookup
	mu    sync.RWMutex
}

// SemanticNode represents a single word in the semantic graph
// Neighbors map stores edges: neighbor ID → similarity weight
type SemanticNode struct {
	word      *dictionary.WordIdentity
	neighbors map[int]float64  // neighbor ID → similarity weight (1.0 for now)
	mu        sync.RWMutex
}

// NewSemanticGraph creates an empty semantic graph
func NewSemanticGraph() *SemanticGraph {
	return &SemanticGraph{
		nodes: make(map[int]*SemanticNode),
		words: make(map[int]*dictionary.WordIdentity),
	}
}

// AddNode adds a word to the graph (idempotent - returns existing if present)
// Creates a node for the word with empty neighbor set
func (g *SemanticGraph) AddNode(word *dictionary.WordIdentity) *SemanticNode {
	if word == nil {
		return nil
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	// Return existing node if already added
	if node, exists := g.nodes[word.ID]; exists {
		return node
	}

	// Create new node
	node := &SemanticNode{
		word:      word,
		neighbors: make(map[int]float64),
	}
	g.nodes[word.ID] = node
	g.words[word.ID] = word
	return node
}

// AddEdge creates a bidirectional edge between two nodes
// Edge weight represents similarity strength (1.0 for co-occurrence)
// Thread-safe for concurrent edge additions
func (g *SemanticGraph) AddEdge(id1, id2 int, similarity float64) {
	if id1 == id2 {
		return // No self-loops
	}

	g.mu.RLock()
	node1 := g.nodes[id1]
	node2 := g.nodes[id2]
	g.mu.RUnlock()

	if node1 != nil && node2 != nil {
		// Add edge in both directions (undirected graph)
		node1.mu.Lock()
		node1.neighbors[id2] = similarity
		node1.mu.Unlock()

		node2.mu.Lock()
		node2.neighbors[id1] = similarity
		node2.mu.Unlock()
	}
}

// GetNode retrieves a node by word ID
func (g *SemanticGraph) GetNode(id int) *SemanticNode {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.nodes[id]
}

// GetWord retrieves a WordIdentity by ID
func (g *SemanticGraph) GetWord(id int) *dictionary.WordIdentity {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.words[id]
}

// NodeCount returns number of nodes in the graph
func (g *SemanticGraph) NodeCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.nodes)
}

// EdgeCount returns total number of edges (counting each edge once)
func (g *SemanticGraph) EdgeCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()

	totalEdges := 0
	for _, node := range g.nodes {
		node.mu.RLock()
		totalEdges += len(node.neighbors)
		node.mu.RUnlock()
	}
	// Each edge is counted twice (bidirectional), so divide by 2
	return totalEdges / 2
}

// GetNeighbors returns all neighbor nodes (for traversal)
func (n *SemanticNode) GetNeighbors() []int {
	n.mu.RLock()
	defer n.mu.RUnlock()

	neighborIDs := make([]int, 0, len(n.neighbors))
	for id := range n.neighbors {
		neighborIDs = append(neighborIDs, id)
	}
	return neighborIDs
}

// NeighborCount returns number of neighbors
func (n *SemanticNode) NeighborCount() int {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return len(n.neighbors)
}

// Word returns the word identity for this node
func (n *SemanticNode) Word() *dictionary.WordIdentity {
	return n.word
}
