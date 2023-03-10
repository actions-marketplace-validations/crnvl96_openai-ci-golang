package chatGPTClient

import (
	"context"
	"fmt"
	"strings"

	gh "github.com/crnvl96/openai-ci-golang/pkg/githubClient"
	"github.com/google/go-github/v50/github"
	gogpt "github.com/sashabaranov/go-gpt3"
)

type GenerateCodeReviewArgs struct {
	GHClient          *github.Client
	GHContext         context.Context
	GPTClient         *gogpt.Client
	GPTContext        context.Context
	RepositoryOwner   string
	RepositoryName    string
	PullRequestNumber int
}

func GenerateCodeReview(args GenerateCodeReviewArgs) {
	commits := gh.RetrieveCommits(
		gh.RetrieveCommitsArgs{
			PullRequestNumber: args.PullRequestNumber,
			GHClient:          args.GHClient,
			GHContext:         args.GHContext,
			RepositoryOwner:   args.RepositoryOwner,
			RepositoryName:    args.RepositoryName,
		},
	)

	lastCommit := commits[len(commits)-1]

	filesFromCommit := gh.RetrieveFiles(
		gh.RetrieveFilesArgs{
			GHClient:        args.GHClient,
			GHContext:       args.GHContext,
			RepositoryOwner: args.RepositoryOwner,
			RepositoryName:  args.RepositoryName,
			CommitSHA:       *lastCommit.SHA,
		},
	)

	for _, file := range filesFromCommit.Files {
		filePath := *file.Filename

		fileContent := gh.RetrieveFileContent(
			gh.RetrieveFileContentArgs{
				GHClient:        args.GHClient,
				GHContext:       args.GHContext,
				RepositoryOwner: args.RepositoryOwner,
				RepositoryName:  args.RepositoryName,
				FilePath:        filePath,
			},
		)

		request := GenerateRequestToGPT(fileContent)

		response := GetCompletion(
			GetCompletionArgs{
				GPTClient:  args.GPTClient,
				GPTContext: args.GPTContext,
				request:    request,
			},
		)

		segmentedFilePath := strings.Split(filePath, "/")
		fileTitle := segmentedFilePath[len(segmentedFilePath)-1]

		gh.GeneratePullRequestComment(
			gh.GeneratePullRequestCommentArgs{
				Body:              fmt.Sprintf("### Code review for ```%s``` \n %s", fileTitle, response),
				GHClient:          args.GHClient,
				GHContext:         args.GHContext,
				RepositoryOwner:   args.RepositoryOwner,
				RepositoryName:    args.RepositoryName,
				PullRequestNumber: args.PullRequestNumber,
			},
		)
	}
}
