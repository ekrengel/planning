package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/go-github/v29/github"
	"golang.org/x/oauth2"
)

func main() {
	token := os.Getenv("GITHUB_AUTH_TOKEN")
	if token == "" {
		log.Fatal("Unauthorized: No token present")
	}
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// Get all repos for the org.
	// repoOpts := github.RepositoryListByOrgOptions{
	// 	ListOptions: github.ListOptions{
	// 		PerPage: 200,
	// 	},
	// }
	// repos, _, err := client.Repositories.ListByOrg(ctx, "pulumi", &repoOpts)
	// if err != nil {
	// 	log.Panicf("error getting repos: %v", err)
	// }

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
			Labels:   []string{"20Q1-svc"},
		}
		issues, _, err := client.Issues.ListByRepo(ctx, "pulumi", repo, opts)
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
		case "size-s":
			return 1
		case "size-m":
			return 5
		case "size-l":
			return 10
		default:
			continue
		}
	}
	// The issue does not have a size label. We will return a high value to
	// make sure we realize something is up.
	return 500
}
