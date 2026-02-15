package spine

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestMarshalFullGraph(t *testing.T) {
	g := NewGraph[string, string](true)
	g.AddNode("a", "alpha")
	g.AddNode("b", "beta")
	g.AddEdge("a", "b", "connects", 1.5)
	g.NodeMeta("a").Set("lang", "go")
	g.NodeMeta("a").SetSchema(Schema{"lang": {Type: FieldString, Required: true}})
	g.EdgeMeta("a", "b").Set("count", 5)

	data, err := Marshal(g, nil)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("parse JSON: %v", err)
	}

	if raw["version"] != float64(1) {
		t.Fatalf("expected version 1, got %v", raw["version"])
	}
	if raw["directed"] != true {
		t.Fatalf("expected directed true, got %v", raw["directed"])
	}

	graph := raw["graph"].(map[string]any)
	nodes := graph["nodes"].([]any)
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}
	edges := graph["edges"].([]any)
	if len(edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(edges))
	}

	meta := raw["metadata"].(map[string]any)
	metaNodes := meta["nodes"].([]any)
	if len(metaNodes) != 1 {
		t.Fatalf("expected 1 meta node, got %d", len(metaNodes))
	}
	metaEdges := meta["edges"].([]any)
	if len(metaEdges) != 1 {
		t.Fatalf("expected 1 meta edge, got %d", len(metaEdges))
	}

	// Check schema is present.
	mn := metaNodes[0].(map[string]any)
	if mn["schema"] == nil {
		t.Fatal("expected schema in node metadata")
	}
}

func TestMarshalGraphOnly(t *testing.T) {
	g := NewGraph[string, string](true)
	g.AddNode("a", "alpha")
	g.NodeMeta("a").Set("k", "v")

	data, err := Marshal(g, &MarshalOptions{Graph: true})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var raw map[string]any
	json.Unmarshal(data, &raw)
	if raw["graph"] == nil {
		t.Fatal("expected graph section")
	}
	if raw["metadata"] != nil {
		t.Fatal("expected no metadata section")
	}
}

func TestMarshalMetaOnly(t *testing.T) {
	g := NewGraph[string, string](true)
	g.AddNode("a", "alpha")
	g.NodeMeta("a").Set("k", "v")

	data, err := Marshal(g, &MarshalOptions{Meta: true})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var raw map[string]any
	json.Unmarshal(data, &raw)
	if raw["graph"] != nil {
		t.Fatal("expected no graph section")
	}
	if raw["metadata"] == nil {
		t.Fatal("expected metadata section")
	}
}

func TestMarshalSubgraph(t *testing.T) {
	g := NewGraph[string, string](true)
	g.AddNode("a", "A")
	g.AddNode("b", "B")
	g.AddNode("c", "C")
	g.AddEdge("a", "b", "ab", 1)
	g.AddEdge("b", "c", "bc", 1)
	g.AddEdge("a", "c", "ac", 1)
	g.NodeMeta("a").Set("k", "va")
	g.NodeMeta("b").Set("k", "vb")
	g.NodeMeta("c").Set("k", "vc")

	data, err := Marshal(g, &MarshalOptions{
		NodeIDs: []string{"a", "b"},
		Graph:   true,
		Meta:    true,
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var snap Snapshot[string, string]
	json.Unmarshal(data, &snap)

	if len(snap.Graph.Nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(snap.Graph.Nodes))
	}
	if len(snap.Graph.Edges) != 1 {
		t.Fatalf("expected 1 edge (a->b), got %d", len(snap.Graph.Edges))
	}
	if len(snap.Meta.Nodes) != 2 {
		t.Fatalf("expected 2 meta nodes, got %d", len(snap.Meta.Nodes))
	}
}

func TestMarshalCompact(t *testing.T) {
	g := NewGraph[string, string](true)
	g.AddNode("a", "A")

	data, err := Marshal(g, &MarshalOptions{Graph: true, Indent: false})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	if bytes.Contains(data, []byte("\n")) {
		t.Fatal("expected compact JSON without newlines")
	}
}

func TestMarshalNilOpts(t *testing.T) {
	g := NewGraph[string, string](true)
	g.AddNode("a", "A")
	g.NodeMeta("a").Set("k", "v")
	g.NodeMeta("a").SetSchema(Schema{"k": {Type: FieldString, Required: true}})

	data, err := Marshal(g, nil)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	// Should be indented.
	if !bytes.Contains(data, []byte("\n")) {
		t.Fatal("expected indented JSON")
	}

	var raw map[string]any
	json.Unmarshal(data, &raw)
	if raw["graph"] == nil {
		t.Fatal("expected graph section")
	}
	if raw["metadata"] == nil {
		t.Fatal("expected metadata section")
	}

	// Check schema is included (nil opts = Schemas: true).
	meta := raw["metadata"].(map[string]any)
	nodes := meta["nodes"].([]any)
	node := nodes[0].(map[string]any)
	if node["schema"] == nil {
		t.Fatal("expected schema with nil opts")
	}
}

func TestMarshalNoSchemas(t *testing.T) {
	g := NewGraph[string, string](true)
	g.AddNode("a", "A")
	g.NodeMeta("a").Set("k", "v")
	g.NodeMeta("a").SetSchema(Schema{"k": {Type: FieldString, Required: true}})

	data, err := Marshal(g, &MarshalOptions{Meta: true, Schemas: false})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var raw map[string]any
	json.Unmarshal(data, &raw)
	meta := raw["metadata"].(map[string]any)
	nodes := meta["nodes"].([]any)
	node := nodes[0].(map[string]any)
	if node["schema"] != nil {
		t.Fatal("expected no schema when Schemas=false")
	}
}

func TestMarshalEmptyGraph(t *testing.T) {
	g := NewGraph[string, string](true)

	data, err := Marshal(g, nil)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var snap Snapshot[string, string]
	if err := json.Unmarshal(data, &snap); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(snap.Graph.Nodes) != 0 {
		t.Fatalf("expected 0 nodes, got %d", len(snap.Graph.Nodes))
	}
	if len(snap.Graph.Edges) != 0 {
		t.Fatalf("expected 0 edges, got %d", len(snap.Graph.Edges))
	}
}

func TestMarshalUndirected(t *testing.T) {
	g := NewGraph[string, string](false)
	g.AddNode("b", "B")
	g.AddNode("a", "A")
	g.AddEdge("b", "a", "ba", 1.0)
	g.EdgeMeta("b", "a").Set("k", "v")

	data, err := Marshal(g, nil)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var snap Snapshot[string, string]
	json.Unmarshal(data, &snap)

	if snap.Directed {
		t.Fatal("expected directed=false")
	}
	if len(snap.Graph.Edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(snap.Graph.Edges))
	}
	// Edge should be normalized: from < to.
	e := snap.Graph.Edges[0]
	if e.From != "a" || e.To != "b" {
		t.Fatalf("expected edge a->b, got %s->%s", e.From, e.To)
	}
	if len(snap.Meta.Edges) != 1 {
		t.Fatalf("expected 1 meta edge, got %d", len(snap.Meta.Edges))
	}
}

func TestUnmarshalRoundTrip(t *testing.T) {
	g := NewGraph[string, string](true)
	g.AddNode("a", "alpha")
	g.AddNode("b", "beta")
	g.AddEdge("a", "b", "connects", 2.5)
	g.NodeMeta("a").Set("lang", "go")
	g.EdgeMeta("a", "b").Set("count", 10)

	data, err := Marshal(g, nil)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	g2, err := Unmarshal[string, string](data)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if g2.Order() != 2 {
		t.Fatalf("expected 2 nodes, got %d", g2.Order())
	}
	if g2.Size() != 1 {
		t.Fatalf("expected 1 edge, got %d", g2.Size())
	}
	n, ok := g2.GetNode("a")
	if !ok || n.Data != "alpha" {
		t.Fatalf("expected node a with data alpha, got %v", n)
	}
	e, ok := g2.GetEdge("a", "b")
	if !ok || e.Data != "connects" || e.Weight != 2.5 {
		t.Fatalf("expected edge a->b with data connects weight 2.5, got %v", e)
	}

	v, ok := g2.NodeMeta("a").Get("lang")
	if !ok || v != "go" {
		t.Fatalf("expected lang=go, got %v", v)
	}
	// Note: int becomes float64 after JSON round-trip.
	cv, ok := g2.EdgeMeta("a", "b").Get("count")
	if !ok {
		t.Fatal("expected count in edge meta")
	}
	if cv != float64(10) {
		t.Fatalf("expected count=10 (float64), got %v (%T)", cv, cv)
	}
}

func TestUnmarshalGraphOnly(t *testing.T) {
	j := `{"version":1,"directed":true,"graph":{"nodes":[{"id":"x","data":"X"}],"edges":[]}}`
	g, err := Unmarshal[string, string]([]byte(j))
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if g.Order() != 1 {
		t.Fatalf("expected 1 node, got %d", g.Order())
	}
	if g.NodeMetaCount("x") != 0 {
		t.Fatal("expected no metadata")
	}
}

func TestUnmarshalWithMeta(t *testing.T) {
	j := `{
		"version": 1,
		"directed": true,
		"graph": {
			"nodes": [{"id": "a", "data": "A"}, {"id": "b", "data": "B"}],
			"edges": [{"from": "a", "to": "b", "data": "e", "weight": 1}]
		},
		"metadata": {
			"nodes": [{"id": "a", "entries": {"k": "v"}}],
			"edges": [{"from": "a", "to": "b", "entries": {"ek": "ev"}}]
		}
	}`
	g, err := Unmarshal[string, string]([]byte(j))
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	v, ok := g.NodeMeta("a").Get("k")
	if !ok || v != "v" {
		t.Fatalf("expected k=v, got %v", v)
	}
	ev, ok := g.EdgeMeta("a", "b").Get("ek")
	if !ok || ev != "ev" {
		t.Fatalf("expected ek=ev, got %v", ev)
	}
}

func TestUnmarshalBadJSON(t *testing.T) {
	_, err := Unmarshal[string, string]([]byte("{bad"))
	if err == nil {
		t.Fatal("expected error for bad JSON")
	}
}

func TestUnmarshalBadVersion(t *testing.T) {
	_, err := Unmarshal[string, string]([]byte(`{"version":99,"directed":true}`))
	if err == nil {
		t.Fatal("expected error for bad version")
	}
}

func TestApplyMeta(t *testing.T) {
	g := NewGraph[string, string](true)
	g.AddNode("a", "A")
	g.AddNode("b", "B")
	g.AddEdge("a", "b", "e", 1)

	j := `{"metadata":{"nodes":[{"id":"a","entries":{"k":"v"}}],"edges":[{"from":"a","to":"b","entries":{"ek":"ev"}}]}}`
	if err := ApplyMeta([]byte(j), g); err != nil {
		t.Fatalf("apply meta: %v", err)
	}

	v, ok := g.NodeMeta("a").Get("k")
	if !ok || v != "v" {
		t.Fatalf("expected k=v, got %v", v)
	}
	ev, ok := g.EdgeMeta("a", "b").Get("ek")
	if !ok || ev != "ev" {
		t.Fatalf("expected ek=ev, got %v", ev)
	}
}

func TestApplyMetaSkipsMissing(t *testing.T) {
	g := NewGraph[string, string](true)
	g.AddNode("a", "A")

	j := `{"metadata":{"nodes":[{"id":"a","entries":{"k":"v"}},{"id":"z","entries":{"k":"v"}}],"edges":[{"from":"a","to":"z","entries":{"k":"v"}}]}}`
	if err := ApplyMeta([]byte(j), g); err != nil {
		t.Fatalf("apply meta: %v", err)
	}

	// Node "a" should have metadata.
	v, ok := g.NodeMeta("a").Get("k")
	if !ok || v != "v" {
		t.Fatalf("expected k=v for node a, got %v", v)
	}
	// Node "z" doesn't exist, should be skipped.
	if g.HasNode("z") {
		t.Fatal("node z should not exist")
	}
}

func TestApplyMetaWithSchema(t *testing.T) {
	g := NewGraph[string, string](true)
	g.AddNode("a", "A")

	j := `{"metadata":{"nodes":[{"id":"a","entries":{"lang":"go"},"schema":{"lang":{"type":"string","required":true}}}],"edges":[]}}`
	if err := ApplyMeta([]byte(j), g); err != nil {
		t.Fatalf("apply meta: %v", err)
	}

	schema := g.NodeMeta("a").GetSchema()
	if schema == nil {
		t.Fatal("expected schema")
	}
	def, ok := schema["lang"]
	if !ok {
		t.Fatal("expected lang in schema")
	}
	if def.Type != FieldString || !def.Required {
		t.Fatalf("unexpected schema def: %+v", def)
	}
}

func TestMarshalDeterministic(t *testing.T) {
	g := NewGraph[string, string](true)
	g.AddNode("c", "C")
	g.AddNode("a", "A")
	g.AddNode("b", "B")
	g.AddEdge("b", "c", "bc", 1)
	g.AddEdge("a", "b", "ab", 2)
	g.AddEdge("a", "c", "ac", 3)
	g.NodeMeta("a").Set("k1", "v1")
	g.NodeMeta("b").Set("k2", "v2")

	d1, err := Marshal(g, nil)
	if err != nil {
		t.Fatalf("marshal 1: %v", err)
	}
	d2, err := Marshal(g, nil)
	if err != nil {
		t.Fatalf("marshal 2: %v", err)
	}

	if !bytes.Equal(d1, d2) {
		t.Fatalf("non-deterministic output:\n%s\nvs\n%s", d1, d2)
	}
}
