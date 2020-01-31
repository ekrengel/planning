package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/google/go-github/v29/github"
	"github.com/jedib0t/go-pretty/table"
	"golang.org/x/oauth2"
)

var (
	flagGitHubOrganization = flag.String("org", "default", "GitHub organization to scan.")
	flagGitHubLabel        = flag.String("label", "default", "GitHub label marking all issues to be included.")
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

	// repos := []string{
	// 	"pulumi-service",
	// 	"pulumi",
	// 	"pulumi-policy",
	// 	"home",
	// 	"docs",
	// 	"pulumi-policy-aws",
	// 	"pulumi-az-pipelines-task",
	// 	"marketing",
	// 	"customer-support",
	// }

	var sumPoints int
	// https://developer.github.com/v3/issues/#list-issues
	opts := &github.IssueListOptions{
		ListOptions: github.ListOptions{
			PerPage: 50,
		},
		Filter: "all",
		State:  "open",
		Labels: []string{*flagGitHubLabel},
	}
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Issue", "Milestone", "Assignee", "Points", "URL"})

	for {
		issues, resp, err := client.Issues.ListByOrg(ctx, *flagGitHubOrganization, opts)
		if err != nil {
			log.Fatalf("error listing GitHub issues: %v", err)
		}

		for _, i := range issues {
			points := getSizeValue(i)
			sumPoints += points

			var milestoneStr string
			if milestone := i.GetMilestone(); milestone != nil {
				milestoneStr = milestone.GetTitle()
			}

			var assigneeStr string
			if assignee := i.GetAssignee(); assignee != nil {
				assigneeStr = assignee.GetName()
			}

			t.AppendRow([]interface{}{i.GetTitle(), milestoneStr, assigneeStr, points, i.GetHTMLURL()})
		}

		// Fecth the next page of results as needed.
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	t.AppendFooter(table.Row{"", "", "Total", sumPoints})
	t.AppendFooter(table.Row{"", "", "Avg Per Milestone", sumPoints / 3.0})
	t.Render()
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
