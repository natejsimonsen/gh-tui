package github

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type MockClient struct {
	dir string
}

func NewMockClient(dir string) (*MockClient, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("mock data dir %q: %w", dir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("mock data path %q is not a directory", dir)
	}
	return &MockClient{dir: dir}, nil
}

func (m *MockClient) ListPRs(_ context.Context, _ Repo, states []string, first int, _ string) (*PRListResult, error) {
	return m.loadPRList(states, first)
}

func (m *MockClient) SearchPRs(_ context.Context, _ Repo, _ string, states []string, first int, _ string) (*PRListResult, error) {
	return m.loadPRList(states, first)
}

func (m *MockClient) SearchReviewRequested(_ context.Context, _ Repo, _ string, states []string, first int, _ string) (*PRListResult, error) {
	return m.loadPRList(states, first)
}

func (m *MockClient) loadPRList(states []string, first int) (*PRListResult, error) {
	var result PRListResult
	if err := m.loadJSON("prs.json", &result); err != nil {
		return nil, err
	}
	if len(states) > 0 {
		allowed := make(map[string]bool, len(states))
		for _, s := range states {
			allowed[strings.ToUpper(s)] = true
		}
		filtered := result.PullRequests[:0]
		for _, pr := range result.PullRequests {
			if allowed[pr.State] {
				filtered = append(filtered, pr)
			}
		}
		result.PullRequests = filtered
		result.TotalCount = len(filtered)
	}
	if first > 0 && first < len(result.PullRequests) {
		result.PullRequests = result.PullRequests[:first]
	}
	return &result, nil
}

func (m *MockClient) GetPRDetail(_ context.Context, _ Repo, number int) (*PRDetail, error) {
	var detail PRDetail
	if err := m.loadJSON(fmt.Sprintf("pr_%d.json", number), &detail); err != nil {
		return nil, err
	}
	return &detail, nil
}

func (m *MockClient) GetPRDiff(_ context.Context, _ Repo, number int) (string, error) {
	b, err := os.ReadFile(filepath.Join(m.dir, fmt.Sprintf("pr_%d.diff", number)))
	if err != nil {
		return "", fmt.Errorf("mock diff for #%d: %w", number, err)
	}
	return string(b), nil
}

func (m *MockClient) SubmitReview(_ context.Context, _ Repo, _ int, _ ReviewInput) error {
	return nil
}

func (m *MockClient) loadJSON(name string, v any) error {
	b, err := os.ReadFile(filepath.Join(m.dir, name))
	if err != nil {
		return fmt.Errorf("mock data %q: %w", name, err)
	}
	return json.Unmarshal(b, v)
}
