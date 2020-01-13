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

	var sumPoints int
	fmt.Printf("\n\nISSUES - POINTS\n")

	opts := &github.IssueListOptions{
		Filter: "all",
		Labels: []string{"20Q1-svc"},
	}
	issues, _, _ := client.Issues.ListByOrg(ctx, "pulumi", opts)
	for _, i := range issues {
		sizePoints := getSizeValue(i)
		sumPoints += sizePoints
		fmt.Printf("%s - %d\n", i.GetTitle(), sizePoints)
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
