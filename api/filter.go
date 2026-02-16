package api

import (
	"fmt"
	"strings"

	"spine"
)

// matchesFilters returns true if the node passes all filters (AND logic).
func matchesFilters(g *spine.Graph[NodeData, EdgeData], nodeID string, filters []MetaFilter) bool {
	if len(filters) == 0 {
		return true
	}
	node, ok := g.GetNode(nodeID)
	if !ok {
		return false
	}
	store := g.NodeMeta(nodeID)
	for _, f := range filters {
		if !matchFilter(store, node.Data, f) {
			return false
		}
	}
	return true
}

// matchFilter evaluates a single filter predicate against a node's structural
// fields and metadata store.
func matchFilter(store *spine.Store, data NodeData, f MetaFilter) bool {
	// Resolve value: structural fields take precedence.
	var val any
	var found bool

	switch f.Key {
	case "status":
		val, found = data.Status, true
	case "label":
		val, found = data.Label, true
	default:
		if store != nil {
			val, found = store.Get(f.Key)
		}
	}

	switch f.Op {
	case "exists":
		return found
	case "eq":
		if !found {
			return false
		}
		return fmt.Sprintf("%v", val) == fmt.Sprintf("%v", f.Value)
	case "neq":
		if !found {
			return true
		}
		return fmt.Sprintf("%v", val) != fmt.Sprintf("%v", f.Value)
	case "contains":
		if !found {
			return false
		}
		return strings.Contains(fmt.Sprintf("%v", val), fmt.Sprintf("%v", f.Value))
	case "gt":
		return compareFloat(val, f.Value, found) > 0
	case "gte":
		return compareFloat(val, f.Value, found) >= 0
	case "lt":
		return compareFloat(val, f.Value, found) < 0
	case "lte":
		return compareFloat(val, f.Value, found) <= 0
	default:
		return false
	}
}

// compareFloat returns -2 if comparison is impossible, otherwise -1, 0, 1.
func compareFloat(a, b any, found bool) int {
	if !found {
		return -2
	}
	af, ok1 := toFloat64(a)
	bf, ok2 := toFloat64(b)
	if !ok1 || !ok2 {
		return -2
	}
	switch {
	case af < bf:
		return -1
	case af > bf:
		return 1
	default:
		return 0
	}
}

func toFloat64(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case int32:
		return float64(n), true
	case json_number:
		f, err := n.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}

// json_number is a type alias to avoid importing encoding/json just for Number.
// JSON unmarshal with Decoder.UseNumber would produce json.Number, but since
// we use map[string]any, numbers arrive as float64 by default.
type json_number = interface{ Float64() (float64, error) }
