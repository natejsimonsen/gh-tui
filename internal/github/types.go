package github

import (
	"context"
	"time"
)

type DataSource interface {
	ListPRs(ctx context.Context, repo Repo, states []string, first int, after string) (*PRListResult, error)
	SearchPRs(ctx context.Context, repo Repo, author string, states []string, first int, after string) (*PRListResult, error)
	SearchReviewRequested(ctx context.Context, repo Repo, user string, states []string, first int, after string) (*PRListResult, error)
	GetPRDetail(ctx context.Context, repo Repo, number int) (*PRDetail, error)
	GetPRDiff(ctx context.Context, repo Repo, number int) (string, error)
	SubmitReview(ctx context.Context, repo Repo, number int, input ReviewInput) error
}

type Repo struct {
	Owner string
	Name  string
}

type PullRequest struct {
	Number    int
	Title     string
	Author    string
	State     string
	IsDraft   bool
	UpdatedAt time.Time
	Labels    []Label
	Reviewers []Reviewer
	CIStatus  string
}

type Label struct {
	Name  string
	Color string
}

type Reviewer struct {
	Login string
	State string
}

type PRDetail struct {
	Number       int
	Title        string
	Body         string
	Author       string
	State        string
	IsDraft      bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
	MergedAt     *time.Time
	ClosedAt     *time.Time
	Additions    int
	Deletions    int
	ChangedFiles int
	BaseBranch   string
	HeadBranch   string
	Mergeable    string
	Labels       []Label
	Assignees    []string
	Reviewers    []Reviewer
	Comments     []Comment
	Reviews      []Review
	Checks       []Check
}

type Comment struct {
	Author    string
	Body      string
	CreatedAt time.Time
}

type Review struct {
	Author      string
	State       string
	Body        string
	SubmittedAt time.Time
	Comments    []ReviewComment
}

type ReviewComment struct {
	Author    string
	Body      string
	Path      string
	Line      int
	CreatedAt time.Time
}

type Check struct {
	Name       string
	Status     string
	Conclusion string
}

type PageInfo struct {
	HasNextPage bool
	EndCursor   string
}

type PRListResult struct {
	PullRequests []PullRequest
	PageInfo     PageInfo
	TotalCount   int
}

type ReviewInput struct {
	Event    string
	Body     string
	Comments []ReviewCommentInput
}

type ReviewCommentInput struct {
	Path string
	Body string
}
