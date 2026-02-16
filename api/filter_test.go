package api

import (
	"spine"
	"testing"
)

func newTestGraph() *spine.Graph[NodeData, EdgeData] {
	g := spine.NewGraph[NodeData, EdgeData](true)
	g.AddNode("a", NodeData{Label: "Alpha", Status: "done"})
	g.AddNode("b", NodeData{Label: "Beta", Status: "pending"})
	g.AddNode("c", NodeData{Label: "Charlie", Status: "running"})
	g.NodeMeta("a").Set("priority", float64(10))
	g.NodeMeta("b").Set("priority", float64(5))
	g.NodeMeta("c").Set("priority", float64(8))
	g.NodeMeta("a").Set("tag", "core")
	return g
}

func TestMatchFilter_Eq(t *testing.T) {
	g := newTestGraph()
	if !matchesFilters(g, "a", []MetaFilter{{Key: "status", Op: "eq", Value: "done"}}) {
		t.Error("expected status=done to match node a")
	}
	if matchesFilters(g, "b", []MetaFilter{{Key: "status", Op: "eq", Value: "done"}}) {
		t.Error("expected status=done to NOT match node b")
	}
}

func TestMatchFilter_Neq(t *testing.T) {
	g := newTestGraph()
	if !matchesFilters(g, "b", []MetaFilter{{Key: "status", Op: "neq", Value: "done"}}) {
		t.Error("expected status!=done to match node b")
	}
	if matchesFilters(g, "a", []MetaFilter{{Key: "status", Op: "neq", Value: "done"}}) {
		t.Error("expected status!=done to NOT match node a")
	}
}

func TestMatchFilter_Contains(t *testing.T) {
	g := newTestGraph()
	if !matchesFilters(g, "a", []MetaFilter{{Key: "label", Op: "contains", Value: "lph"}}) {
		t.Error("expected label contains 'lph' to match Alpha")
	}
	if matchesFilters(g, "b", []MetaFilter{{Key: "label", Op: "contains", Value: "lph"}}) {
		t.Error("expected label contains 'lph' to NOT match Beta")
	}
}

func TestMatchFilter_Numeric(t *testing.T) {
	g := newTestGraph()
	tests := []struct {
		nodeID string
		op     string
		value  float64
		want   bool
	}{
		{"a", "gt", 9, true},
		{"a", "gt", 10, false},
		{"b", "lt", 10, true},
		{"b", "gte", 5, true},
		{"b", "lte", 4, false},
	}
	for _, tt := range tests {
		got := matchesFilters(g, tt.nodeID, []MetaFilter{{Key: "priority", Op: tt.op, Value: tt.value}})
		if got != tt.want {
			t.Errorf("node %s priority %s %v: got %v, want %v", tt.nodeID, tt.op, tt.value, got, tt.want)
		}
	}
}

func TestMatchFilter_Exists(t *testing.T) {
	g := newTestGraph()
	if !matchesFilters(g, "a", []MetaFilter{{Key: "tag", Op: "exists"}}) {
		t.Error("expected tag exists to match node a")
	}
	if matchesFilters(g, "b", []MetaFilter{{Key: "tag", Op: "exists"}}) {
		t.Error("expected tag exists to NOT match node b")
	}
}

func TestMatchFilter_AND(t *testing.T) {
	g := newTestGraph()
	filters := []MetaFilter{
		{Key: "status", Op: "eq", Value: "done"},
		{Key: "priority", Op: "gte", Value: float64(10)},
	}
	if !matchesFilters(g, "a", filters) {
		t.Error("expected AND filter to match node a")
	}
	if matchesFilters(g, "b", filters) {
		t.Error("expected AND filter to NOT match node b")
	}
}

func TestMatchFilter_MissingNode(t *testing.T) {
	g := newTestGraph()
	if matchesFilters(g, "zzz", []MetaFilter{{Key: "status", Op: "eq", Value: "done"}}) {
		t.Error("expected filter on missing node to return false")
	}
}

func TestMatchFilter_UnknownOp(t *testing.T) {
	g := newTestGraph()
	if matchesFilters(g, "a", []MetaFilter{{Key: "status", Op: "regex", Value: ".*"}}) {
		t.Error("expected unknown op to return false")
	}
}
