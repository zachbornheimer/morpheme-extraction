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
// Positions tracks where this word appears in the ngram array for fast lookup
type SemanticNode struct {
	word      *dictionary.WordIdentity
	neighbors map[int]float64  // neighbor ID → similarity weight (1.0 for now)
	positions []int             // ngram array indices where this word appears
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
// Edge weight represents similarity strength - increments with each co-occurrence
// This implements "neural pathway strengthening" where repeated evidence builds weight
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
		// Increment weight for each occurrence (neural pathway strengthening)
		node1.mu.Lock()
		node1.neighbors[id2] += similarity
		node1.mu.Unlock()

		node2.mu.Lock()
		node2.neighbors[id1] += similarity
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

// GetEdgeWeight returns the weight of the edge to a neighbor
// Returns 0.0 if neighbor doesn't exist
func (n *SemanticNode) GetEdgeWeight(neighborID int) float64 {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.neighbors[neighborID]
}

// AddPosition adds an ngram array index where this word appears
// This enables O(1) lookup instead of O(n) scanning during morpheme extraction
func (n *SemanticNode) AddPosition(ngramIndex int) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.positions = append(n.positions, ngramIndex)
}

// GetPositions returns all ngram array indices where this word appears
// Used for fast lookup during morpheme pair processing
func (n *SemanticNode) GetPositions() []int {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.positions
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

// GetAllNodes returns all nodes in the graph (for export/iteration)
func (g *SemanticGraph) GetAllNodes() []*SemanticNode {
	g.mu.RLock()
	defer g.mu.RUnlock()

	nodes := make([]*SemanticNode, 0, len(g.nodes))
	for _, node := range g.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}

// GetNeighborsWithWeights returns all neighbors with their edge weights
func (n *SemanticNode) GetNeighborsWithWeights() map[int]float64 {
	n.mu.RLock()
	defer n.mu.RUnlock()

	// Return a copy to avoid concurrent modification issues
	neighbors := make(map[int]float64, len(n.neighbors))
	for id, weight := range n.neighbors {
		neighbors[id] = weight
	}
	return neighbors
}

// FindPath uses BFS to find the shortest path between two nodes
// Returns nil if no path exists
// Returns slice of nodes representing the path from fromID to toID (inclusive)
func (g *SemanticGraph) FindPath(fromID, toID int) []*SemanticNode {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.nodes[fromID] == nil || g.nodes[toID] == nil {
		return nil
	}

	if fromID == toID {
		return []*SemanticNode{g.nodes[fromID]}
	}

	// BFS implementation
	visited := make(map[int]bool)
	parent := make(map[int]int)
	queue := []int{fromID}
	visited[fromID] = true

	// BFS search
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current == toID {
			// Reconstruct path by following parent pointers
			path := []*SemanticNode{}
			for id := toID; id != fromID; id = parent[id] {
				path = append([]*SemanticNode{g.nodes[id]}, path...)
			}
			path = append([]*SemanticNode{g.nodes[fromID]}, path...)
			return path
		}

		// Explore neighbors
		node := g.nodes[current]
		node.mu.RLock()
		for neighborID := range node.neighbors {
			if !visited[neighborID] {
				visited[neighborID] = true
				parent[neighborID] = current
				queue = append(queue, neighborID)
			}
		}
		node.mu.RUnlock()
	}

	return nil // No path found
}

// GetCluster returns all nodes within depth hops of the given node
// Uses BFS to explore the graph up to the specified depth
// Returns nil if node doesn't exist
func (g *SemanticGraph) GetCluster(nodeID int, depth int) []*SemanticNode {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.nodes[nodeID] == nil {
		return nil
	}

	if depth < 0 {
		return nil
	}

	if depth == 0 {
		return []*SemanticNode{g.nodes[nodeID]}
	}

	cluster := []*SemanticNode{}
	visited := make(map[int]bool)

	type queueItem struct {
		id    int
		level int
	}

	queue := []queueItem{{nodeID, 0}}
	visited[nodeID] = true

	// BFS with depth tracking
	for len(queue) > 0 {
		item := queue[0]
		queue = queue[1:]

		cluster = append(cluster, g.nodes[item.id])

		// Only explore neighbors if we haven't reached max depth
		if item.level < depth {
			node := g.nodes[item.id]
			node.mu.RLock()
			for neighborID := range node.neighbors {
				if !visited[neighborID] {
					visited[neighborID] = true
					queue = append(queue, queueItem{neighborID, item.level + 1})
				}
			}
			node.mu.RUnlock()
		}
	}

	return cluster
}
