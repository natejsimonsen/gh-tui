package github

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func setupClient(t *testing.T) (*Client, Repo) {
	t.Helper()
	client, err := NewClient()
	if err != nil {
		t.Skipf("no GitHub token available: %v", err)
	}
	return client, Repo{Owner: "cli", Name: "cli"}
}

func TestAPILatency_ListPRs(t *testing.T) {
	client, repo := setupClient(t)

	start := time.Now()
	result, err := client.ListPRs(context.Background(), repo, []string{"OPEN"}, 20, "")
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("ListPRs failed: %v", err)
	}

	t.Logf("ListPRs: %d PRs in %v", len(result.PullRequests), elapsed)

	if elapsed > 5*time.Second {
		t.Errorf("ListPRs too slow: %v", elapsed)
	}
}

func TestAPILatency_SinglePRDetail(t *testing.T) {
	client, repo := setupClient(t)

	result, err := client.ListPRs(context.Background(), repo, []string{"OPEN"}, 1, "")
	if err != nil {
		t.Fatalf("ListPRs failed: %v", err)
	}
	if len(result.PullRequests) == 0 {
		t.Skip("no open PRs")
	}

	pr := result.PullRequests[0]
	start := time.Now()
	detail, err := client.GetPRDetail(context.Background(), repo, pr.Number)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("GetPRDetail failed: %v", err)
	}

	t.Logf("GetPRDetail #%d: %v (comments=%d reviews=%d checks=%d)",
		detail.Number, elapsed, len(detail.Comments), len(detail.Reviews), len(detail.Checks))

	if elapsed > 5*time.Second {
		t.Errorf("GetPRDetail too slow: %v", elapsed)
	}
}

func TestAPILatency_ConcurrentPRDetails(t *testing.T) {
	client, repo := setupClient(t)

	result, err := client.ListPRs(context.Background(), repo, []string{"OPEN"}, 15, "")
	if err != nil {
		t.Fatalf("ListPRs failed: %v", err)
	}
	if len(result.PullRequests) == 0 {
		t.Skip("no open PRs")
	}

	for _, concurrency := range []int{1, 3, 5, 10, 15} {
		count := concurrency
		if count > len(result.PullRequests) {
			count = len(result.PullRequests)
		}

		t.Run(fmt.Sprintf("concurrent_%d", count), func(t *testing.T) {
			prs := result.PullRequests[:count]
			sem := make(chan struct{}, count)
			var wg sync.WaitGroup
			var mu sync.Mutex
			var durations []time.Duration

			overall := time.Now()

			for _, pr := range prs {
				wg.Add(1)
				sem <- struct{}{}
				go func(number int) {
					defer wg.Done()
					defer func() { <-sem }()

					start := time.Now()
					_, err := client.GetPRDetail(context.Background(), repo, number)
					d := time.Since(start)

					mu.Lock()
					durations = append(durations, d)
					mu.Unlock()

					if err != nil {
						t.Errorf("GetPRDetail #%d failed: %v", number, err)
					}
				}(pr.Number)
			}

			wg.Wait()
			total := time.Since(overall)

			var sum time.Duration
			var maxD time.Duration
			for _, d := range durations {
				sum += d
				if d > maxD {
					maxD = d
				}
			}
			avg := sum / time.Duration(len(durations))

			t.Logf("concurrency=%d total=%v avg=%v max=%v", count, total, avg, maxD)
		})
	}
}

func TestAPILatency_PRDiff(t *testing.T) {
	client, repo := setupClient(t)

	result, err := client.ListPRs(context.Background(), repo, []string{"OPEN"}, 1, "")
	if err != nil {
		t.Fatalf("ListPRs failed: %v", err)
	}
	if len(result.PullRequests) == 0 {
		t.Skip("no open PRs")
	}

	pr := result.PullRequests[0]
	start := time.Now()
	diff, err := client.GetPRDiff(context.Background(), repo, pr.Number)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("GetPRDiff failed: %v", err)
	}

	t.Logf("GetPRDiff #%d: %v (%d bytes)", pr.Number, elapsed, len(diff))

	if elapsed > 5*time.Second {
		t.Errorf("GetPRDiff too slow: %v", elapsed)
	}
}
