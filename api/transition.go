package api

import "fmt"

// validTransitions defines allowed status changes.
var validTransitions = map[string]map[string]bool{
	"":        {"pending": true, "ready": true},
	"pending": {"ready": true, "skipped": true},
	"ready":   {"running": true, "skipped": true},
	"running": {"done": true, "failed": true},
	"failed":  {"pending": true},
}

// Transition moves a node to a new status, enforcing valid transitions.
// When a node becomes "done", downstream nodes whose deps are all done
// are automatically promoted to "ready".
func (m *Manager) Transition(req TransitionRequest) (*TransitionResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	g, err := m.getGraph(req.Graph)
	if err != nil {
		return nil, err
	}

	node, ok := g.GetNode(req.ID)
	if !ok {
		return nil, fmt.Errorf("node %q not found", req.ID)
	}

	oldStatus := node.Data.Status
	newStatus := req.Status

	allowed, exists := validTransitions[oldStatus]
	if !exists || !allowed[newStatus] {
		return nil, fmt.Errorf("invalid transition: %q -> %q", oldStatus, newStatus)
	}

	// Apply the transition.
	nd := node.Data
	nd.Status = newStatus
	g.AddNode(req.ID, nd)

	res := &TransitionResult{
		ID:        req.ID,
		OldStatus: oldStatus,
		NewStatus: newStatus,
	}

	// Auto-ready propagation: when status becomes "done", check downstream.
	if newStatus == "done" {
		for _, outEdge := range g.OutEdges(req.ID) {
			downstream, ok := g.GetNode(outEdge.To)
			if !ok || downstream.Data.Status != "pending" {
				continue
			}
			// Check if ALL incoming deps are done.
			allDone := true
			for _, inEdge := range g.InEdges(outEdge.To) {
				dep, ok := g.GetNode(inEdge.From)
				if !ok || dep.Data.Status != "done" {
					allDone = false
					break
				}
			}
			if allDone {
				dd := downstream.Data
				dd.Status = "ready"
				g.AddNode(outEdge.To, dd)
				res.NewlyReady = append(res.NewlyReady, outEdge.To)
			}
		}
	}

	return res, nil
}
