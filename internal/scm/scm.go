// Package scm provides a generic source control management interface
// (GitHub via gh CLI, GitLab via glab CLI).
package scm

import (
	"context"
	"fmt"
)

// SCM abstracts source control operations.
type SCM interface {
	// CreatePR creates a pull/merge request. Returns the PR URL.
	CreatePR(ctx context.Context, title, body, branch string) (string, error)

	// GetIssue fetches an issue by number.
	GetIssue(ctx context.Context, number int) (*Issue, error)

	// PostComment adds a comment to an issue/PR.
	PostComment(ctx context.Context, issueNumber int, body string) error

	// ListLabels returns labels on an issue.
	ListLabels(ctx context.Context, issueNumber int) ([]string, error)
}

// Issue represents a GitHub/GitLab issue.
type Issue struct {
	Number    int      `json:"number"`
	Title     string   `json:"title"`
	Body      string   `json:"body"`
	Labels    []string `json:"labels"`
	Assignee  string   `json:"assignee"`
	Milestone string   `json:"milestone"`
	State     string   `json:"state"`
}

// Resolve detects the SCM backend from the current git remote.
func Resolve(ctx context.Context) (SCM, error) {
	// TODO: detect github.com vs gitlab.com from git remote URL
	return nil, fmt.Errorf("scm.Resolve: not yet implemented")
}
