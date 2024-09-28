package gofacto

type dag struct {
	nodes map[string]assocNode
	edges map[string][]string
}

func newDAG() *dag {
	return &dag{
		nodes: make(map[string]assocNode),
		edges: make(map[string][]string),
	}
}

func (d *dag) addNode(node assocNode) {
	d.nodes[node.name] = node
}

func (d *dag) addEdge(from, to string) {
	d.edges[from] = append(d.edges[from], to)
}

// topologicalSort returns the topological sort of the DAG
func (d *dag) topologicalSort() []assocNode {
	visited := make(map[string]bool)
	result := make([]assocNode, len(d.nodes))
	i := len(result) - 1

	var dfs func(node string)
	dfs = func(node string) {
		if visited[node] {
			return
		}

		visited[node] = true
		for _, neighbor := range d.edges[node] {
			dfs(neighbor)
		}

		result[i] = d.nodes[node]
		i--
	}

	for node := range d.nodes {
		dfs(node)
	}

	return result
}

// hasCycle returns true if the DAG has a cycle
func (d *dag) hasCycle() bool {
	visited := make(map[string]bool)
	recursionStack := make(map[string]bool)

	var dfs func(node string) bool
	dfs = func(node string) bool {
		if recursionStack[node] {
			return true
		}

		if visited[node] {
			return false
		}

		visited[node] = true
		recursionStack[node] = true
		for _, neighbor := range d.edges[node] {
			if dfs(neighbor) {
				return true
			}
		}

		recursionStack[node] = false
		return false
	}

	for node := range d.nodes {
		if dfs(node) {
			return true
		}
	}

	return false
}
