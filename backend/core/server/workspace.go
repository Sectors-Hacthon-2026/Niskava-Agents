// Package server provides the background REST/SSE daemon server for Niskava Agent.
package server

import (
	_ "embed"
	"fmt"
	"strings"
)

//go:embed workspace.html
var workspaceHTMLTemplate string

// RenderWorkspaceHTML renders the complete, responsive dual-theme Web Workspace.
func RenderWorkspaceHTML(port int) string {
	return strings.ReplaceAll(workspaceHTMLTemplate, "{{PORT}}", fmt.Sprintf("%d", port))
}
