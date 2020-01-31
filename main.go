package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/google/go-github/v29/github"
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

	repos := []string{
		"pulumi-service",
		"pulumi",
		"pulumi-policy",
		"home",
		"docs",
		"pulumi-policy-aws",
		"pulumi-az-pipelines-task",
		"marketing",
		"customer-support",
	}

	var sumPoints int
	fmt.Printf("\n\nISSUES - POINTS\n")

	for _, repo := range repos {
		opts := &github.IssueListByRepoOptions{
			Assignee: "*",
			Labels:   []string{*flagGitHubLabel},
		}
		issues, _, err := client.Issues.ListByRepo(ctx, *flagGitHubOrganization, repo, opts)
		if err != nil {
			log.Panicf("error getting issues: %v", err)
		}
		for _, i := range issues {
			sizePoints := getSizeValue(i)
			sumPoints += sizePoints
			fmt.Printf("%s - %d\n", i.GetTitle(), sizePoints)
		}

	}

	fmt.Printf("\n\nTOTAL SUM: %d\n", sumPoints)
	fmt.Printf("AVG PER MILESTONE: %d\n", sumPoints/3.0)
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
