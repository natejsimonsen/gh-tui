package github

const listPRsQuery = `
query($owner: String!, $name: String!, $states: [PullRequestState!], $first: Int!, $after: String) {
  repository(owner: $owner, name: $name) {
    pullRequests(states: $states, first: $first, after: $after, orderBy: {field: UPDATED_AT, direction: DESC}) {
      totalCount
      pageInfo {
        hasNextPage
        endCursor
      }
      nodes {
        number
        title
        author { login }
        state
        isDraft
        updatedAt
        labels(first: 5) {
          nodes { name color }
        }
        reviewRequests(first: 10) {
          nodes {
            requestedReviewer {
              ... on User { login }
              ... on Team { name }
            }
          }
        }
        latestOpinionatedReviews(first: 10) {
          nodes {
            author { login }
            state
          }
        }
        commits(last: 1) {
          nodes {
            commit {
              statusCheckRollup {
                state
              }
            }
          }
        }
      }
    }
  }
}
`

const prDetailQuery = `
query($owner: String!, $name: String!, $number: Int!) {
  repository(owner: $owner, name: $name) {
    pullRequest(number: $number) {
      number
      title
      body
      author { login }
      state
      isDraft
      createdAt
      updatedAt
      mergedAt
      closedAt
      additions
      deletions
      changedFiles
      baseRefName
      headRefName
      mergeable
      labels(first: 20) {
        nodes { name color }
      }
      assignees(first: 10) {
        nodes { login }
      }
      reviewRequests(first: 10) {
        nodes {
          requestedReviewer {
            ... on User { login }
            ... on Team { name }
          }
        }
      }
      reviews(last: 30) {
        nodes {
          author { login }
          state
          body
          submittedAt
          comments(first: 50) {
            nodes {
              author { login }
              body
              path
              position
              createdAt
            }
          }
        }
      }
      commits(last: 1) {
        nodes {
          commit {
            statusCheckRollup {
              state
              contexts(first: 100) {
                nodes {
                  ... on CheckRun {
                    __typename
                    name
                    status
                    conclusion
                  }
                  ... on StatusContext {
                    __typename
                    context
                    state
                  }
                }
              }
            }
          }
        }
      }
      comments(first: 100) {
        nodes {
          author { login }
          body
          createdAt
        }
      }
    }
  }
}
`
