# Neo4j Semantic Graph Setup

This document explains how to explore the semantic graph using Neo4j.

## Quick Start

1. **Start Neo4j in Docker:**
   ```bash
   docker-compose up -d
   ```

2. **Access Neo4j Browser:**
   - Open: http://localhost:7474
   - Username: `neo4j`
   - Password: `password`

3. **Import the semantic graph:**
   ```bash
   # The graph file is quite large (413MB for 100KB corpus)
   # Load it in the Neo4j browser or use cypher-shell
   docker exec -it morpheme-neo4j cypher-shell -u neo4j -p password
   ```

   Then in the cypher-shell:
   ```cypher
   :source /import/semantic-graph.cypher
   ```

## Graph Structure

- **Nodes**: Words (WordIdentity)
  - `id`: Unique identifier
  - `text`: The word string
  - `centerFreq`: Frequency as center word in n-grams

- **Edges**: RELATED relationships
  - `weight`: Semantic/syntactic/morphological strength (2.0 to 10,612.0)
  - Higher weight = stronger relationship (more co-occurrences in similar contexts)

## Example Queries

### Find a specific word node:
```cypher
MATCH (w:Word {text: 'walk'}) RETURN w;
```

### Find strongest relationships for a word:
```cypher
MATCH (w:Word {text: 'walk'})-[r:RELATED]-(neighbor:Word)
RETURN neighbor.text, r.weight
ORDER BY r.weight DESC
LIMIT 20;
```

### Find semantic clusters (words highly interconnected):
```cypher
MATCH (w:Word)-[r:RELATED]->(neighbor:Word)
WHERE r.weight > 100
WITH w, COUNT(neighbor) as degreeCount
RETURN w.text, degreeCount
ORDER BY degreeCount DESC
LIMIT 50;
```

### Find path between two words:
```cypher
MATCH p=shortestPath((a:Word {text: 'walk'})-[*]-(b:Word {text: 'ocean'}))
RETURN p;
```

### Find words with similar neighbors (potential lexical category):
```cypher
MATCH (w1:Word)-[r1:RELATED]->(shared:Word)<-[r2:RELATED]-(w2:Word)
WHERE w1.text = 'walk' AND w1 <> w2
WITH w2, COUNT(shared) as sharedNeighbors
RETURN w2.text, sharedNeighbors
ORDER BY sharedNeighbors DESC
LIMIT 20;
```

### Find high-degree nodes (potential category centers):
```cypher
MATCH (w:Word)-[r:RELATED]-()
WITH w, COUNT(r) as degree
RETURN w.text, degree, w.centerFreq
ORDER BY degree DESC
LIMIT 50;
```

## Performance Notes

- The initial load of 4.8M relationships takes several minutes
- Create indexes (already in the .cypher file):
  ```cypher
  CREATE INDEX word_id_index IF NOT EXISTS FOR (w:Word) ON (w.id);
  CREATE INDEX word_text_index IF NOT EXISTS FOR (w:Word) ON (w.text);
  ```

- For large graphs, consider using Neo4j's bulk import tools instead

## Troubleshooting

If the import times out or fails:
1. Increase Neo4j memory in `docker-compose.yml`
2. Use Neo4j's `neo4j-admin import` for bulk loading
3. Or import in batches using `CALL apoc.periodic.iterate()`

## Stopping Neo4j

```bash
docker-compose down
```

To completely remove data:
```bash
docker-compose down -v
rm -rf neo4j-data neo4j-logs
```
