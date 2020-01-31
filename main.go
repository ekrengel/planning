package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/google/go-github/v29/github"
	"github.com/jedib0t/go-pretty/table"
	"golang.org/x/oauth2"
)

// Issue contains the information for a GitHub issue that we care about for planning purposes.
type Issue struct {
	Issue     string
	Milestone string
	Assignee  string
	Points    int
	URL       string
}

var (
	flagGitHubOrganization = flag.String("org", "default", "GitHub organization to scan.")
	flagGitHubLabel        = flag.String("label", "default", "GitHub label marking all issues to be included.")
	flagAll                = flag.Bool("all", false, "Prints all issues in one table")
)

func main() {
	flag.Parse()

	token := os.Getenv("GITHUB_AUTH_TOKEN")
	if token == "" {
		log.Fatal("Unauthorized: No token present")
	}

	if flagGitHubOrganization == nil || *flagGitHubOrganization == "" {
		log.Fatal("Error: Required --org flag not proided.")
	}
	if flagGitHubLabel == nil || *flagGitHubLabel == "" {
		log.Fatalf("Error: Required --label flag not provided.")
	}

	log.Printf(
		"Scanning GitHub organization %q and all issues labeled %q...",
		*flagGitHubOrganization, *flagGitHubLabel)

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// https://developer.github.com/v3/issues/#list-issues
	opts := &github.IssueListOptions{
		ListOptions: github.ListOptions{
			PerPage: 50,
		},
		Filter: "all",
		State:  "open",
		Labels: []string{*flagGitHubLabel},
	}
	var allIssues []Issue

	for {
		issues, resp, err := client.Issues.ListByOrg(ctx, *flagGitHubOrganization, opts)
		if err != nil {
			log.Fatalf("error listing GitHub issues: %v", err)
		}

		for _, i := range issues {
			allIssues = append(allIssues, createIssue(i))
		}

		// Fetch the next page of results as needed.
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	// We only print the all issues view, if the all flag was specified.
	if flagAll != nil && *flagAll {
		renderTable(allIssues, "All Issues")
		return
	}

	renderPerAssignee(allIssues)
}

func createIssue(gitHubIssue *github.Issue) Issue {
	points := getSizeValue(gitHubIssue)

	var milestoneStr string
	if milestone := gitHubIssue.GetMilestone(); milestone != nil {
		milestoneStr = milestone.GetTitle()
	}

	var assigneeStr string
	if assignee := gitHubIssue.GetAssignee(); assignee != nil {
		assigneeStr = assignee.GetLogin()
	}
	return Issue{
		Issue:     gitHubIssue.GetTitle(),
		Milestone: milestoneStr,
		Assignee:  assigneeStr,
		Points:    points,
		URL:       gitHubIssue.GetHTMLURL(),
	}
}

func getSizeValue(issue *github.Issue) int {
	for _, l := range issue.Labels {
		switch size := l.GetName(); size {
		case "size-s", "size/S":
			return 1
		case "size-m", "size/M":
			return 5
		case "size-l", "size/L":
			return 10
		default:
			continue
		}
	}
	// The issue does not have a size label. We will return a high value to
	// make sure we realize something is up.
	return 500
}

func renderTable(issues []Issue, headerText string) {
	pointsPerMilestone := make(map[string]int)

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{headerText})
	t.AppendHeader(table.Row{"Issue", "Milestone", "Assignee", "Points", "URL"})
	var sumPoints int
	for _, i := range issues {
		sumPoints += i.Points
		t.AppendRow([]interface{}{i.Issue, i.Milestone, i.Assignee, i.Points, i.URL})
		pointsPerMilestone[i.Milestone] += i.Points
	}

	t.AppendFooter(table.Row{"", "", "Total", sumPoints})

	for milestone, points := range pointsPerMilestone {
		milestoneText := fmt.Sprintf("Points for %s", milestone)
		t.AppendFooter(table.Row{"", "", milestoneText, points})
	}
	t.Render()
}

func renderPerAssignee(issues []Issue) {
	issuesByAssignee := make(map[string][]Issue)

	for _, i := range issues {
		if val, ok := issuesByAssignee[i.Assignee]; ok {
			issuesByAssignee[i.Assignee] = append(val, i)
		} else {
			issuesByAssignee[i.Assignee] = []Issue{i}
		}
	}

	for assignee, assignedIssues := range issuesByAssignee {
		renderTable(assignedIssues, fmt.Sprintf("Issues for %s", assignee))
	}
}
