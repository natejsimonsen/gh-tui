package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Client struct {
	httpClient *http.Client
	token      string
}

func NewClient() (*Client, error) {
	token := detectToken()
	if token == "" {
		return nil, fmt.Errorf("no GitHub token found: run 'gh auth login' or set GITHUB_TOKEN")
	}
	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		token:      token,
	}, nil
}

func detectToken() string {
	out, err := exec.Command("gh", "auth", "token").Output()
	if err == nil {
		if t := strings.TrimSpace(string(out)); t != "" {
			return t
		}
	}
	return os.Getenv("GITHUB_TOKEN")
}

type graphQLRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

type graphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func (c *Client) query(ctx context.Context, q string, variables map[string]any) (json.RawMessage, error) {
	b, err := json.Marshal(graphQLRequest{Query: q, Variables: variables})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.github.com/graphql", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var gqlResp graphQLResponse
	if err := json.Unmarshal(body, &gqlResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	if len(gqlResp.Errors) > 0 {
		return nil, fmt.Errorf("graphql: %s", gqlResp.Errors[0].Message)
	}

	return gqlResp.Data, nil
}

type actor struct {
	Login string `json:"login"`
}

type reviewerNode struct {
	Login string `json:"login"`
	Name  string `json:"name"`
}

// List PRs

type listPRsResponse struct {
	Repository struct {
		PullRequests struct {
			TotalCount int `json:"totalCount"`
			PageInfo   struct {
				HasNextPage bool   `json:"hasNextPage"`
				EndCursor   string `json:"endCursor"`
			} `json:"pageInfo"`
			Nodes []prNode `json:"nodes"`
		} `json:"pullRequests"`
	} `json:"repository"`
}

type prNode struct {
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	Author    *actor    `json:"author"`
	State     string    `json:"state"`
	IsDraft   bool      `json:"isDraft"`
	UpdatedAt time.Time `json:"updatedAt"`
	Labels    struct {
		Nodes []struct {
			Name  string `json:"name"`
			Color string `json:"color"`
		} `json:"nodes"`
	} `json:"labels"`
	ReviewRequests struct {
		Nodes []struct {
			RequestedReviewer reviewerNode `json:"requestedReviewer"`
		} `json:"nodes"`
	} `json:"reviewRequests"`
	LatestOpinionatedReviews struct {
		Nodes []struct {
			Author *actor `json:"author"`
			State  string `json:"state"`
		} `json:"nodes"`
	} `json:"latestOpinionatedReviews"`
	Commits struct {
		Nodes []struct {
			Commit struct {
				StatusCheckRollup *struct {
					State string `json:"state"`
				} `json:"statusCheckRollup"`
			} `json:"commit"`
		} `json:"nodes"`
	} `json:"commits"`
}

func (c *Client) ListPRs(ctx context.Context, repo Repo, states []string, first int, after string) (*PRListResult, error) {
	vars := map[string]any{
		"owner": repo.Owner,
		"name":  repo.Name,
		"first": first,
	}
	if len(states) > 0 {
		vars["states"] = states
	}
	if after != "" {
		vars["after"] = after
	}

	data, err := c.query(ctx, listPRsQuery, vars)
	if err != nil {
		return nil, err
	}

	var resp listPRsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	result := &PRListResult{
		TotalCount: resp.Repository.PullRequests.TotalCount,
		PageInfo: PageInfo{
			HasNextPage: resp.Repository.PullRequests.PageInfo.HasNextPage,
			EndCursor:   resp.Repository.PullRequests.PageInfo.EndCursor,
		},
	}

	for _, n := range resp.Repository.PullRequests.Nodes {
		pr := PullRequest{
			Number:    n.Number,
			Title:     n.Title,
			Author:    actorLogin(n.Author),
			State:     n.State,
			IsDraft:   n.IsDraft,
			UpdatedAt: n.UpdatedAt,
		}

		for _, l := range n.Labels.Nodes {
			pr.Labels = append(pr.Labels, Label{Name: l.Name, Color: l.Color})
		}

		seen := map[string]bool{}
		for _, r := range n.LatestOpinionatedReviews.Nodes {
			login := actorLogin(r.Author)
			if login != "" {
				pr.Reviewers = append(pr.Reviewers, Reviewer{Login: login, State: r.State})
				seen[login] = true
			}
		}
		for _, r := range n.ReviewRequests.Nodes {
			login := r.RequestedReviewer.Login
			if login == "" {
				login = r.RequestedReviewer.Name
			}
			if login != "" && !seen[login] {
				pr.Reviewers = append(pr.Reviewers, Reviewer{Login: login, State: "REQUESTED"})
			}
		}

		if len(n.Commits.Nodes) > 0 {
			if rollup := n.Commits.Nodes[0].Commit.StatusCheckRollup; rollup != nil {
				pr.CIStatus = rollup.State
			}
		}

		result.PullRequests = append(result.PullRequests, pr)
	}

	return result, nil
}

// PR Detail

type prDetailResponse struct {
	Repository struct {
		PullRequest prDetailNode `json:"pullRequest"`
	} `json:"repository"`
}

type prDetailNode struct {
	Number       int        `json:"number"`
	Title        string     `json:"title"`
	Body         string     `json:"body"`
	Author       *actor     `json:"author"`
	State        string     `json:"state"`
	IsDraft      bool       `json:"isDraft"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
	MergedAt     *time.Time `json:"mergedAt"`
	ClosedAt     *time.Time `json:"closedAt"`
	Additions    int        `json:"additions"`
	Deletions    int        `json:"deletions"`
	ChangedFiles int        `json:"changedFiles"`
	BaseRefName  string     `json:"baseRefName"`
	HeadRefName  string     `json:"headRefName"`
	Mergeable    string     `json:"mergeable"`
	Labels       struct {
		Nodes []struct {
			Name  string `json:"name"`
			Color string `json:"color"`
		} `json:"nodes"`
	} `json:"labels"`
	Assignees struct {
		Nodes []struct {
			Login string `json:"login"`
		} `json:"nodes"`
	} `json:"assignees"`
	ReviewRequests struct {
		Nodes []struct {
			RequestedReviewer reviewerNode `json:"requestedReviewer"`
		} `json:"nodes"`
	} `json:"reviewRequests"`
	Reviews struct {
		Nodes []struct {
			Author      *actor    `json:"author"`
			State       string    `json:"state"`
			Body        string    `json:"body"`
			SubmittedAt time.Time `json:"submittedAt"`
			Comments    struct {
				Nodes []struct {
					Author    *actor    `json:"author"`
					Body      string    `json:"body"`
					Path      string    `json:"path"`
					Position  *int      `json:"position"`
					CreatedAt time.Time `json:"createdAt"`
				} `json:"nodes"`
			} `json:"comments"`
		} `json:"nodes"`
	} `json:"reviews"`
	Commits struct {
		Nodes []struct {
			Commit struct {
				StatusCheckRollup *struct {
					State    string `json:"state"`
					Contexts struct {
						Nodes []json.RawMessage `json:"nodes"`
					} `json:"contexts"`
				} `json:"statusCheckRollup"`
			} `json:"commit"`
		} `json:"nodes"`
	} `json:"commits"`
	Comments struct {
		Nodes []struct {
			Author    *actor    `json:"author"`
			Body      string    `json:"body"`
			CreatedAt time.Time `json:"createdAt"`
		} `json:"nodes"`
	} `json:"comments"`
}

type checkRunNode struct {
	TypeName   string `json:"__typename"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
}

type statusContextNode struct {
	TypeName string `json:"__typename"`
	Context  string `json:"context"`
	State    string `json:"state"`
}

func (c *Client) GetPRDetail(ctx context.Context, repo Repo, number int) (*PRDetail, error) {
	vars := map[string]any{
		"owner":  repo.Owner,
		"name":   repo.Name,
		"number": number,
	}

	data, err := c.query(ctx, prDetailQuery, vars)
	if err != nil {
		return nil, err
	}

	var resp prDetailResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	n := resp.Repository.PullRequest
	detail := &PRDetail{
		Number:       n.Number,
		Title:        n.Title,
		Body:         n.Body,
		Author:       actorLogin(n.Author),
		State:        n.State,
		IsDraft:      n.IsDraft,
		CreatedAt:    n.CreatedAt,
		UpdatedAt:    n.UpdatedAt,
		MergedAt:     n.MergedAt,
		ClosedAt:     n.ClosedAt,
		Additions:    n.Additions,
		Deletions:    n.Deletions,
		ChangedFiles: n.ChangedFiles,
		BaseBranch:   n.BaseRefName,
		HeadBranch:   n.HeadRefName,
		Mergeable:    n.Mergeable,
	}

	for _, l := range n.Labels.Nodes {
		detail.Labels = append(detail.Labels, Label{Name: l.Name, Color: l.Color})
	}
	for _, a := range n.Assignees.Nodes {
		detail.Assignees = append(detail.Assignees, a.Login)
	}

	seen := map[string]bool{}
	for _, r := range n.Reviews.Nodes {
		login := actorLogin(r.Author)
		if login != "" && !seen[login] {
			detail.Reviewers = append(detail.Reviewers, Reviewer{Login: login, State: r.State})
			seen[login] = true
		}
	}
	for _, r := range n.ReviewRequests.Nodes {
		login := r.RequestedReviewer.Login
		if login == "" {
			login = r.RequestedReviewer.Name
		}
		if login != "" && !seen[login] {
			detail.Reviewers = append(detail.Reviewers, Reviewer{Login: login, State: "REQUESTED"})
		}
	}

	for _, c := range n.Comments.Nodes {
		detail.Comments = append(detail.Comments, Comment{
			Author:    actorLogin(c.Author),
			Body:      c.Body,
			CreatedAt: c.CreatedAt,
		})
	}

	for _, r := range n.Reviews.Nodes {
		review := Review{
			Author:      actorLogin(r.Author),
			State:       r.State,
			Body:        r.Body,
			SubmittedAt: r.SubmittedAt,
		}
		for _, rc := range r.Comments.Nodes {
			line := 0
			if rc.Position != nil {
				line = *rc.Position
			}
			review.Comments = append(review.Comments, ReviewComment{
				Author:    actorLogin(rc.Author),
				Body:      rc.Body,
				Path:      rc.Path,
				Line:      line,
				CreatedAt: rc.CreatedAt,
			})
		}
		detail.Reviews = append(detail.Reviews, review)
	}

	if len(n.Commits.Nodes) > 0 {
		if rollup := n.Commits.Nodes[0].Commit.StatusCheckRollup; rollup != nil {
			for _, raw := range rollup.Contexts.Nodes {
				var peek struct {
					TypeName string `json:"__typename"`
				}
				if err := json.Unmarshal(raw, &peek); err != nil {
					continue
				}
				switch peek.TypeName {
				case "CheckRun":
					var cr checkRunNode
					json.Unmarshal(raw, &cr)
					detail.Checks = append(detail.Checks, Check{
						Name:       cr.Name,
						Status:     cr.Status,
						Conclusion: cr.Conclusion,
					})
				case "StatusContext":
					var sc statusContextNode
					json.Unmarshal(raw, &sc)
					detail.Checks = append(detail.Checks, Check{
						Name:       sc.Context,
						Status:     sc.State,
						Conclusion: sc.State,
					})
				}
			}
		}
	}

	return detail, nil
}

func ParseRepo(s string) (Repo, error) {
	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return Repo{}, fmt.Errorf("invalid repo format %q, expected owner/repo", s)
	}
	return Repo{Owner: parts[0], Name: parts[1]}, nil
}

func DetectRepo() (Repo, error) {
	out, err := exec.Command("git", "remote", "get-url", "origin").Output()
	if err != nil {
		return Repo{}, fmt.Errorf("not in a git repo or no origin remote: %w", err)
	}
	url := strings.TrimSpace(string(out))
	url = strings.TrimSuffix(url, ".git")

	if strings.HasPrefix(url, "git@github.com:") {
		return ParseRepo(strings.TrimPrefix(url, "git@github.com:"))
	}
	if idx := strings.Index(url, "github.com/"); idx >= 0 {
		return ParseRepo(url[idx+len("github.com/"):])
	}

	return Repo{}, fmt.Errorf("could not parse GitHub repo from remote URL: %s", url)
}

func actorLogin(a *actor) string {
	if a == nil {
		return "ghost"
	}
	return a.Login
}
