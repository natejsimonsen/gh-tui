package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/natejsimonsen/gh-tui/internal/github"
)

var users = []string{
	"alice", "bob", "charlie", "diana", "eve", "frank",
	"grace", "heidi", "ivan", "judy", "karl", "liam",
	"mallory", "nina", "oscar", "peggy", "quinn", "rupert",
}

var labels = []github.Label{
	{Name: "bug", Color: "d73a4a"},
	{Name: "enhancement", Color: "a2eeef"},
	{Name: "documentation", Color: "0075ca"},
	{Name: "refactor", Color: "cfd3d7"},
	{Name: "performance", Color: "f9d0c4"},
	{Name: "security", Color: "e4e669"},
	{Name: "dependencies", Color: "0366d6"},
	{Name: "breaking-change", Color: "b60205"},
	{Name: "WIP", Color: "fbca04"},
	{Name: "needs-review", Color: "7057ff"},
}

var titles = []string{
	"Fix race condition in cache invalidation",
	"Add retry logic for transient API failures",
	"Refactor authentication middleware",
	"Migrate user sessions to Redis",
	"Update OpenAPI spec for v3 endpoints",
	"Implement rate limiting for public routes",
	"Fix memory leak in WebSocket handler",
	"Add structured logging with trace IDs",
	"Optimize database queries for dashboard",
	"Replace deprecated crypto package",
	"Add dark mode support to settings page",
	"Fix pagination offset-by-one error",
	"Implement RBAC for admin endpoints",
	"Add health check endpoint for k8s probes",
	"Refactor event bus to use channels",
	"Fix timezone handling in scheduler",
	"Add CSV export for analytics reports",
	"Implement graceful shutdown for workers",
	"Fix XSS vulnerability in comment renderer",
	"Add integration tests for payment flow",
	"Migrate to Go 1.22 iterator pattern",
	"Fix deadlock in connection pool",
	"Add OpenTelemetry tracing",
	"Refactor config loading to use envconfig",
	"Fix flaky test in notification service",
	"Add bulk delete endpoint for resources",
	"Implement circuit breaker for downstream calls",
	"Fix incorrect Content-Type on file uploads",
	"Add database migration for audit log",
	"Refactor error handling to use error types",
	"Fix CORS headers for preflight requests",
	"Add webhook signature verification",
	"Implement streaming response for large exports",
	"Fix N+1 query in team members endpoint",
	"Add request ID middleware",
	"Refactor job scheduler to use cron syntax",
	"Fix race in concurrent map access",
	"Add feature flags service integration",
	"Implement soft delete for resources",
	"Fix OAuth callback URL mismatch",
	"Add compression for API responses",
	"Refactor file storage to support S3",
	"Fix cache stampede on cold start",
	"Add E2E tests for onboarding flow",
	"Implement API versioning headers",
	"Fix goroutine leak in SSE handler",
	"Add Prometheus metrics for queue depth",
	"Refactor middleware chain ordering",
	"Fix TLS certificate rotation",
	"Add batch processing for notifications",
}

var bodies = []string{
	"## Summary\nThis PR addresses a long-standing issue where concurrent cache invalidation could leave stale entries. The fix uses a compare-and-swap approach to ensure atomicity.\n\n## Changes\n- Added `sync.Map` for thread-safe cache operations\n- Replaced manual locking with CAS-based updates\n- Added benchmark tests showing 3x improvement\n\n## Testing\n- Unit tests for concurrent access patterns\n- Load test with 1000 concurrent readers/writers\n- Verified no regression in p99 latency",
	"## What\nAdds exponential backoff with jitter for all external API calls. Previously, transient 503s from upstream would cascade into user-facing errors.\n\n## Why\nWe've seen 15+ incidents in the last quarter caused by brief upstream outages that our system amplified instead of absorbing.\n\n## How\n- New `retry` package with configurable backoff\n- Applied to HTTP client, gRPC client, and database connections\n- Circuit breaker integration to avoid retry storms\n\n## Metrics\nExpect to see a ~40% reduction in 5xx errors during partial outages.",
	"This is a large refactor of the auth middleware stack. The existing code had grown organically over 2 years and had several issues:\n\n1. Token validation was duplicated in 3 places\n2. Session refresh logic was inconsistent\n3. No support for API keys alongside JWT\n\nThe new design uses a chain-of-responsibility pattern where each auth method is a pluggable handler.\n\n### Migration\nNo breaking changes to the API. Internal interfaces have changed — see updated docs in `/docs/auth.md`.\n\n### Risks\n- Auth is critical path — deploying behind feature flag\n- Canary for 24h before full rollout\n- Rollback plan documented in deploy ticket",
}

var reviewBodies = []string{
	"Looks good overall. A few minor suggestions but nothing blocking.",
	"Nice approach! I like the separation of concerns here.",
	"LGTM - clean implementation. Just one nit about naming.",
	"I have some concerns about the error handling approach. Let's discuss.",
	"Great work on the tests. Coverage looks solid.",
	"",
	"",
}

var commentBodies = []string{
	"Can we add a timeout here? If the upstream service is slow, this will block the entire request.",
	"This looks like it could panic on nil. Should we add a guard?",
	"Nice catch! I missed this edge case in my review of the earlier PR.",
	"Have we considered using `sync.Pool` here instead? Might reduce GC pressure.",
	"Nit: this variable name is a bit misleading — it's not actually a count, it's an index.",
	"Should this be behind a feature flag for the initial rollout?",
	"Can we add a test for the error case? The happy path is covered but I don't see failure scenarios.",
	"I think this introduces a subtle race condition. Consider using `atomic.Value` instead.",
	"This is going to be O(n²) for large inputs. Can we use a map lookup instead?",
	"The error message here isn't very helpful for debugging. Can we include the request ID?",
	"+1, this is much cleaner than the old approach.",
	"Wouldn't `context.WithTimeout` be more appropriate here than `context.WithDeadline`?",
	"This breaks backward compatibility with v2 clients. We need a migration path.",
	"Can we extract this into a helper? I see the same pattern in 3 other files.",
	"The SQL here is vulnerable to injection if the input isn't sanitized upstream. Worth adding a check.",
	"Good use of `errgroup` here — much cleaner than manual goroutine management.",
	"This log line will be very noisy in production. Should we make it debug-level?",
	"Does this handle the case where the slice is empty? I think it'll panic on line 42.",
	"We should add an index on this column before deploying — the query plan shows a full table scan.",
	"Consider using `strings.Builder` instead of `fmt.Sprintf` in the hot path.",
}

var inlineCommentBodies = []string{
	"This should use `errors.Is` instead of string comparison.",
	"Missing `defer rows.Close()` — this will leak connections.",
	"The mutex scope is too wide here. Can we narrow it to just the critical section?",
	"This constant should probably be configurable via environment variable.",
	"Unused import — was this from a previous iteration?",
	"Consider using a named return here for clarity in the error path.",
	"This nil check is redundant — the constructor guarantees non-nil.",
	"Off-by-one: should be `< len(items)` not `<= len(items)`.",
	"Can we use `time.Since(start)` instead of `time.Now().Sub(start)`?",
	"This goroutine has no way to be cancelled. Pass the context through.",
}

var filePaths = []string{
	"internal/auth/middleware.go",
	"internal/auth/token.go",
	"internal/cache/redis.go",
	"internal/cache/memory.go",
	"internal/config/config.go",
	"internal/db/migrations/0042_add_audit_log.sql",
	"internal/db/queries.go",
	"internal/handler/api.go",
	"internal/handler/webhook.go",
	"internal/model/user.go",
	"internal/model/team.go",
	"internal/service/notification.go",
	"internal/service/scheduler.go",
	"internal/service/export.go",
	"pkg/retry/backoff.go",
	"pkg/retry/circuit.go",
	"cmd/server/main.go",
	"cmd/worker/main.go",
	"test/integration/auth_test.go",
	"test/integration/api_test.go",
	"docs/api.md",
	"Dockerfile",
	"docker-compose.yml",
	".github/workflows/ci.yml",
}

var checkNames = []string{
	"ci/build", "ci/lint", "ci/test-unit", "ci/test-integration",
	"ci/test-e2e", "security/snyk", "security/codeql", "deploy/preview",
	"coverage/codecov", "ci/typecheck", "ci/docker-build", "license/check",
}

func main() {
	outDir := flag.String("out", "testdata/mock", "output directory")
	prCount := flag.Int("prs", 50, "number of PRs to generate")
	seed := flag.Int64("seed", 42, "random seed for reproducibility")
	flag.Parse()

	rng := rand.New(rand.NewSource(*seed))

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	now := time.Now()
	prs := make([]github.PullRequest, *prCount)
	for i := range prs {
		prs[i] = genPR(rng, i+1, now)
	}

	result := github.PRListResult{
		PullRequests: prs,
		TotalCount:   len(prs),
		PageInfo:     github.PageInfo{HasNextPage: false},
	}
	writeJSON(filepath.Join(*outDir, "prs.json"), result)

	for _, pr := range prs {
		detail := genDetail(rng, pr, now)
		writeJSON(filepath.Join(*outDir, fmt.Sprintf("pr_%d.json", pr.Number)), detail)

		diff := genDiff(rng, detail)
		os.WriteFile(filepath.Join(*outDir, fmt.Sprintf("pr_%d.diff", pr.Number)), []byte(diff), 0o644)
	}

	fmt.Printf("Generated mock data for %d PRs in %s\n", *prCount, *outDir)
}

func genPR(rng *rand.Rand, number int, now time.Time) github.PullRequest {
	states := []string{"OPEN", "OPEN", "OPEN", "MERGED", "CLOSED"}
	state := states[rng.Intn(len(states))]

	pr := github.PullRequest{
		Number:    number,
		Title:     titles[rng.Intn(len(titles))],
		Author:    users[rng.Intn(len(users))],
		State:     state,
		IsDraft:   state == "OPEN" && rng.Float64() < 0.15,
		UpdatedAt: now.Add(-time.Duration(rng.Intn(30*24)) * time.Hour),
	}

	labelCount := rng.Intn(4)
	used := map[int]bool{}
	for j := 0; j < labelCount; j++ {
		idx := rng.Intn(len(labels))
		if !used[idx] {
			pr.Labels = append(pr.Labels, labels[idx])
			used[idx] = true
		}
	}

	reviewerCount := 1 + rng.Intn(4)
	usedReviewer := map[string]bool{pr.Author: true}
	reviewStates := []string{"APPROVED", "CHANGES_REQUESTED", "COMMENTED", "REQUESTED"}
	for j := 0; j < reviewerCount; j++ {
		u := users[rng.Intn(len(users))]
		if !usedReviewer[u] {
			pr.Reviewers = append(pr.Reviewers, github.Reviewer{
				Login: u,
				State: reviewStates[rng.Intn(len(reviewStates))],
			})
			usedReviewer[u] = true
		}
	}

	ciStates := []string{"SUCCESS", "SUCCESS", "SUCCESS", "FAILURE", "PENDING"}
	pr.CIStatus = ciStates[rng.Intn(len(ciStates))]

	return pr
}

func genDetail(rng *rand.Rand, pr github.PullRequest, now time.Time) github.PRDetail {
	created := pr.UpdatedAt.Add(-time.Duration(1+rng.Intn(14*24)) * time.Hour)

	detail := github.PRDetail{
		Number:       pr.Number,
		Title:        pr.Title,
		Body:         bodies[rng.Intn(len(bodies))],
		Author:       pr.Author,
		State:        pr.State,
		IsDraft:      pr.IsDraft,
		CreatedAt:    created,
		UpdatedAt:    pr.UpdatedAt,
		Additions:    200 + rng.Intn(1200),
		Deletions:    50 + rng.Intn(600),
		ChangedFiles: 3 + rng.Intn(25),
		BaseBranch:   "main",
		HeadBranch:   fmt.Sprintf("feat/pr-%d", pr.Number),
		Mergeable:    "MERGEABLE",
		Labels:       pr.Labels,
		Reviewers:    pr.Reviewers,
	}

	if pr.State == "MERGED" {
		t := pr.UpdatedAt
		detail.MergedAt = &t
	} else if pr.State == "CLOSED" {
		t := pr.UpdatedAt
		detail.ClosedAt = &t
	}

	assigneeCount := rng.Intn(3)
	for i := 0; i < assigneeCount; i++ {
		detail.Assignees = append(detail.Assignees, users[rng.Intn(len(users))])
	}

	commentCount := 3 + rng.Intn(20)
	for i := 0; i < commentCount; i++ {
		detail.Comments = append(detail.Comments, github.Comment{
			Author:    users[rng.Intn(len(users))],
			Body:      commentBodies[rng.Intn(len(commentBodies))],
			CreatedAt: created.Add(time.Duration(i) * time.Hour),
		})
	}

	reviewCount := 2 + rng.Intn(6)
	for i := 0; i < reviewCount; i++ {
		reviewStates := []string{"APPROVED", "CHANGES_REQUESTED", "COMMENTED"}
		review := github.Review{
			Author:      users[rng.Intn(len(users))],
			State:       reviewStates[rng.Intn(len(reviewStates))],
			Body:        reviewBodies[rng.Intn(len(reviewBodies))],
			SubmittedAt: created.Add(time.Duration(1+i*3) * time.Hour),
		}

		inlineCount := rng.Intn(8)
		for j := 0; j < inlineCount; j++ {
			review.Comments = append(review.Comments, github.ReviewComment{
				Author:    review.Author,
				Body:      inlineCommentBodies[rng.Intn(len(inlineCommentBodies))],
				Path:      filePaths[rng.Intn(len(filePaths))],
				Line:      1 + rng.Intn(200),
				CreatedAt: review.SubmittedAt.Add(time.Duration(j) * time.Minute),
			})
		}

		detail.Reviews = append(detail.Reviews, review)
	}

	checkCount := 4 + rng.Intn(9)
	for i := 0; i < checkCount; i++ {
		conclusion := []string{"SUCCESS", "SUCCESS", "SUCCESS", "SUCCESS", "FAILURE", "NEUTRAL", "SKIPPED"}
		detail.Checks = append(detail.Checks, github.Check{
			Name:       checkNames[i%len(checkNames)],
			Status:     "COMPLETED",
			Conclusion: conclusion[rng.Intn(len(conclusion))],
		})
	}

	_ = now
	return detail
}

func genDiff(rng *rand.Rand, detail github.PRDetail) string {
	var b strings.Builder
	fileCount := detail.ChangedFiles
	if fileCount > len(filePaths) {
		fileCount = len(filePaths)
	}

	linesPerFile := (detail.Additions + detail.Deletions) / max(fileCount, 1)

	perm := rng.Perm(len(filePaths))
	for i := 0; i < fileCount; i++ {
		path := filePaths[perm[i]]
		fmt.Fprintf(&b, "diff --git a/%s b/%s\n", path, path)
		fmt.Fprintf(&b, "index %07x..%07x 100644\n", rng.Intn(0xFFFFFFF), rng.Intn(0xFFFFFFF))
		fmt.Fprintf(&b, "--- a/%s\n", path)
		fmt.Fprintf(&b, "+++ b/%s\n", path)

		hunkCount := 1 + rng.Intn(5)
		startLine := 1
		linesThisFile := linesPerFile + rng.Intn(max(linesPerFile/2, 1))

		for h := 0; h < hunkCount; h++ {
			hunkLines := linesThisFile/hunkCount + rng.Intn(max(linesThisFile/(hunkCount*2), 1))
			startLine += rng.Intn(40) + 5
			fmt.Fprintf(&b, "@@ -%d,%d +%d,%d @@\n", startLine, hunkLines, startLine, hunkLines+rng.Intn(10)-5)

			for l := 0; l < hunkLines; l++ {
				roll := rng.Float64()
				switch {
				case roll < 0.3:
					b.WriteString("+" + genCodeLine(rng, path) + "\n")
				case roll < 0.5:
					b.WriteString("-" + genCodeLine(rng, path) + "\n")
				default:
					b.WriteString(" " + genCodeLine(rng, path) + "\n")
				}
			}
			startLine += hunkLines
		}
	}

	return b.String()
}

func genCodeLine(rng *rand.Rand, path string) string {
	if strings.HasSuffix(path, ".go") {
		return goLines[rng.Intn(len(goLines))]
	}
	if strings.HasSuffix(path, ".sql") {
		return sqlLines[rng.Intn(len(sqlLines))]
	}
	if strings.HasSuffix(path, ".yml") || strings.HasSuffix(path, ".yaml") {
		return yamlLines[rng.Intn(len(yamlLines))]
	}
	if strings.HasSuffix(path, ".md") {
		return mdLines[rng.Intn(len(mdLines))]
	}
	return goLines[rng.Intn(len(goLines))]
}

var goLines = []string{
	"\tctx, cancel := context.WithTimeout(ctx, 30*time.Second)",
	"\tdefer cancel()",
	"\tif err != nil {",
	"\t\treturn fmt.Errorf(\"failed to connect: %w\", err)",
	"\t}",
	"\tresult, err := db.QueryContext(ctx, query, args...)",
	"\tdefer result.Close()",
	"\tfor result.Next() {",
	"\t\tvar row Record",
	"\t\tif err := result.Scan(&row.ID, &row.Name, &row.Value); err != nil {",
	"\t\t\treturn nil, err",
	"\t\t}",
	"\t\trecords = append(records, row)",
	"\tmu.Lock()",
	"\tdefer mu.Unlock()",
	"\tcache[key] = &entry{value: v, expiry: time.Now().Add(ttl)}",
	"\tgo func() {",
	"\t\tselect {",
	"\t\tcase <-ctx.Done():",
	"\t\t\treturn",
	"\t\tcase msg := <-ch:",
	"\t\t\thandler.Process(msg)",
	"\t\t}",
	"\t}()",
	"\tresp, err := http.DefaultClient.Do(req)",
	"\tif resp.StatusCode != http.StatusOK {",
	"\t\treturn fmt.Errorf(\"unexpected status: %d\", resp.StatusCode)",
	"\tlog.Printf(\"processing %d items\", len(items))",
	"\tslog.Info(\"request completed\", \"duration\", time.Since(start), \"status\", resp.StatusCode)",
	"\treturn json.NewEncoder(w).Encode(response)",
	"\tw.Header().Set(\"Content-Type\", \"application/json\")",
	"\tw.WriteHeader(http.StatusCreated)",
	"\ttoken, err := jwt.Parse(raw, keyFunc)",
	"\tclaims, ok := token.Claims.(jwt.MapClaims)",
	"\tif !ok || !token.Valid {",
	"\t\thttp.Error(w, \"unauthorized\", http.StatusUnauthorized)",
	"\tvar wg sync.WaitGroup",
	"\twg.Add(len(tasks))",
	"\terrCh := make(chan error, len(tasks))",
	"\tfor _, task := range tasks {",
	"\t\tgo func(t Task) {",
	"\t\t\tdefer wg.Done()",
	"\t\t\tif err := t.Execute(ctx); err != nil {",
	"\t\t\t\terrCh <- err",
	"\t\t\t}",
	"\t\t}(task)",
	"\twg.Wait()",
	"\tclose(errCh)",
	"func (s *Server) handleHealthCheck(w http.ResponseWriter, r *http.Request) {",
	"func (s *Server) Shutdown(ctx context.Context) error {",
	"\ts.listener.Close()",
	"\treturn s.server.Shutdown(ctx)",
	"type Config struct {",
	"\tDatabaseURL string `env:\"DATABASE_URL\" required:\"true\"`",
	"\tRedisURL    string `env:\"REDIS_URL\" default:\"localhost:6379\"`",
	"\tPort        int    `env:\"PORT\" default:\"8080\"`",
	"\tDebug       bool   `env:\"DEBUG\" default:\"false\"`",
}

var sqlLines = []string{
	"CREATE TABLE IF NOT EXISTS audit_log (",
	"  id BIGSERIAL PRIMARY KEY,",
	"  user_id BIGINT NOT NULL REFERENCES users(id),",
	"  action VARCHAR(255) NOT NULL,",
	"  resource_type VARCHAR(100) NOT NULL,",
	"  resource_id BIGINT,",
	"  metadata JSONB DEFAULT '{}',",
	"  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()",
	");",
	"CREATE INDEX idx_audit_log_user ON audit_log(user_id);",
	"CREATE INDEX idx_audit_log_created ON audit_log(created_at);",
	"ALTER TABLE users ADD COLUMN last_login_at TIMESTAMPTZ;",
	"SELECT u.id, u.name, COUNT(p.id) as pr_count",
	"FROM users u LEFT JOIN pull_requests p ON u.id = p.author_id",
	"WHERE u.active = true",
	"GROUP BY u.id, u.name",
	"ORDER BY pr_count DESC;",
}

var yamlLines = []string{
	"name: CI",
	"on: [push, pull_request]",
	"jobs:",
	"  build:",
	"    runs-on: ubuntu-latest",
	"    steps:",
	"      - uses: actions/checkout@v4",
	"      - uses: actions/setup-go@v5",
	"        with:",
	"          go-version: '1.22'",
	"      - run: go build ./...",
	"      - run: go test -race ./...",
	"    services:",
	"      postgres:",
	"        image: postgres:16",
	"        env:",
	"          POSTGRES_DB: testdb",
	"          POSTGRES_PASSWORD: test",
}

var mdLines = []string{
	"# API Reference",
	"## Authentication",
	"All requests require a Bearer token in the Authorization header.",
	"### POST /api/v2/resources",
	"Creates a new resource.",
	"| Field | Type | Required |",
	"|-------|------|----------|",
	"| name | string | yes |",
	"| type | string | no |",
	"```json",
	"{\"status\": \"ok\", \"data\": [...]}",
	"```",
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func writeJSON(path string, v any) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling %s: %s\n", path, err)
		os.Exit(1)
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing %s: %s\n", path, err)
		os.Exit(1)
	}
}
