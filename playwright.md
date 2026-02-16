# Playwright MCP Skill - Development Guide for Spine

This document describes how to use the Playwright MCP browser skill to develop and test the Spine Graph Visualizer.

## Overview

The Playwright MCP skill provides browser automation via an accessibility-first approach. It can navigate to URLs, take screenshots, interact with DOM elements, and run arbitrary Playwright code. This is useful for visually verifying UI changes, testing interactions, and debugging the Spine visualizer.

## Starting the Visualizer

```bash
# Kill any existing process on port 8090
fuser -k 8090/tcp 2>/dev/null

# Start the server (from the spine directory)
cd /home/dev/spine && make run
```

The server runs at `http://localhost:8090`.

**Note:** The `make run` target uses `go run ./cmd/visualizer/`. The binary prints the "running" message before binding the port, so a bind error means something else is already on 8090. Use `netstat -tlnp | grep 8090` to find the PID if `fuser` is unavailable.

## Key Playwright MCP Tools

### Navigation

```
browser_navigate  - Go to a URL (e.g., http://localhost:8090)
browser_navigate_back - Go back in history
browser_close - Close the browser page
```

### Inspecting the Page

```
browser_snapshot - Get an accessibility tree of the page (preferred for interactions)
browser_take_screenshot - Capture a visual screenshot (preferred for verifying styling/colors)
browser_console_messages - View browser console output (useful for debugging JS errors)
browser_network_requests - View network requests
```

### Interacting with Elements

All interactions use `ref` values from `browser_snapshot` output:

```
browser_click - Click an element by ref
browser_type - Type text into an input
browser_select_option - Select a dropdown option
browser_fill_form - Fill multiple form fields at once
browser_press_key - Press a keyboard key (e.g., Escape, Enter)
browser_hover - Hover over an element
browser_drag - Drag and drop between elements
browser_file_upload - Upload files
```

### Advanced

```
browser_run_code - Execute arbitrary Playwright JavaScript
browser_evaluate - Evaluate JS on the page or an element
browser_wait_for - Wait for text/conditions
browser_tabs - Manage browser tabs
browser_resize - Resize the browser window
```

## Spine UI Structure

The visualizer is a single-page app with these main areas:

### Sidebar (left panel)
- **Graph Mode** - Toggle directed/undirected
- **Import/Export** - JSON import/export buttons
- **Directory Upload** - Upload a local directory as a file tree graph
- **Templates** - Dropdown to load pre-built graphs (combobox `ref`)
- **Add Node/Edge** - Manual graph construction
- **Metadata Panel** - Appears when a node/edge is selected, shows key-value metadata entries
- **Algorithms** - BFS, DFS, Shortest Path, Topo Sort, etc.

### Canvas (main area)
- HTML `<canvas>` element - nodes and edges are rendered here
- Nodes are **not** DOM elements; they must be clicked via coordinates using `browser_run_code`

### Fullscreen Metadata Viewer (modal overlay)
- Opens when clicking a metadata entry in the sidebar
- Shows content with line numbers and syntax highlighting
- Close via Escape key or the X button

## Common Workflows

### 1. Load a Template and Verify

```
1. browser_navigate to http://localhost:8090
2. browser_select_option on the template dropdown
3. browser_click the "Load Template" button
4. browser_take_screenshot to verify the graph rendered
```

### 2. Click a Canvas Node

Canvas elements don't have DOM refs. Use coordinate-based clicks:

```
browser_run_code with:
  async (page) => {
    await page.mouse.click(362, 168);  // x, y coordinates
    await page.waitForTimeout(500);
  }
```

**Tip:** Take a screenshot first to estimate node positions from the visual layout.

### 3. Inspect Metadata

After selecting a node (via canvas click):
```
1. browser_snapshot - Find the metadata panel entries in the accessibility tree
2. browser_click on a metadata entry ref to open the fullscreen viewer
3. browser_take_screenshot to verify syntax highlighting, formatting, etc.
```

### 4. Check Syntax Highlighting

The fullscreen metadata viewer uses highlight.js (loaded from CDN) for code files. Language detection works by:
1. Checking the metadata key for a file extension (e.g., `main.go` -> Go)
2. Falling back to the selected node's ID for a file extension (e.g., node `graph.go` with key `content` -> Go)
3. JSON objects use the built-in JSON highlighter
4. Unrecognized content shows as plain text

To verify highlighting works:
```
1. Load the "Codebase File Tree" template
2. Click a .go node on the canvas
3. Click the "content" metadata entry
4. Take a screenshot - code should have colored syntax tokens
```

### 5. Debug JavaScript Errors

```
browser_console_messages with level: "error"
```

### 6. Test Form Interactions

```
browser_fill_form with fields array, e.g.:
  [{"name": "Node ID", "type": "textbox", "ref": "e40", "value": "myNode"}]
```

## Tips

- **Snapshot vs Screenshot**: Use `browser_snapshot` when you need to find element refs for interaction. Use `browser_take_screenshot` when you need to verify visual appearance (colors, layout, highlighting).
- **Canvas interactions**: The graph canvas doesn't expose accessible elements for nodes/edges. Always use `browser_run_code` with `page.mouse.click(x, y)` for canvas interactions.
- **Refs change**: Element refs from snapshots are invalidated after page state changes. Always take a fresh snapshot before interacting if the page has changed.
- **Wait for state**: After clicking buttons that trigger API calls, use `browser_wait_for` or add `page.waitForTimeout()` in `browser_run_code` before taking snapshots/screenshots.
- **Server restarts**: After code changes to the Go backend or static HTML, kill the server (`fuser -k 8090/tcp`) and restart with `make run`. The frontend is served from `cmd/visualizer/static/index.html` which is embedded at compile time, so Go changes require a rebuild.
