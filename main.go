package main

import (
	"context"
	"log"
	"os"
	"sync"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

func main() {
	ctx := context.Background()

	token, ok := os.LookupEnv("GITHUB_TOKEN")
	if !ok {
		log.Fatal("GITHUB_TOKEN environment variable is not set")
	}

	orgName, ok := os.LookupEnv("GITHUB_ORG_NAME")
	if !ok {
		log.Fatal("GITHUB_ORG_NAME environment variable is not set")
	}

	tokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: token,
		},
	)
	httpClient := oauth2.NewClient(ctx, tokenSource)
	client := github.NewClient(httpClient)

	var wg sync.WaitGroup
	opt := github.ListOptions{
		PerPage: 100,
	}

	for {
		repos, resp, err := client.Activity.ListWatched(ctx, "", &opt)
		if err != nil {
			log.Fatalf("Error listing watched repositories: %s", err)
		}

		for _, repo := range repos {
			if *repo.Owner.Login == orgName {
				wg.Add(1)
				go func(repoName string) {
					defer wg.Done()

					if _, err := client.Activity.DeleteRepositorySubscription(ctx, orgName, repoName); err != nil {
						log.Printf("Error unsubscribing from repository %s: %s", repoName, err)
					} else {
						log.Printf("Unsubscribed from repository %s", repoName)
					}
				}(*repo.Name)
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	wg.Wait() // Wait for all goroutines to finish
}
