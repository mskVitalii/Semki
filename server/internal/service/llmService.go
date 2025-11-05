package service

import (
	"context"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"semki/internal/model"
)

type ILLMService interface {
	DescribeUser(ctx context.Context, query string, org model.Organization, user model.User) (string, error)
}

type LLMService struct {
	client *openai.Client
}

func NewLLMService(openAIKey string) ILLMService {
	client := openai.NewClient(openAIKey)
	return &LLMService{client: client}
}

func (s *LLMService) DescribeUser(ctx context.Context, query string, org model.Organization, user model.User) (string, error) {
	teamName := ""
	levelName := ""
	locationName := user.Semantic.Location.Hex()

	for _, t := range org.Semantic.Teams {
		if t.ID == user.Semantic.Team {
			teamName = t.Name
			break
		}
	}

	for _, l := range org.Semantic.Levels {
		if l.ID == user.Semantic.Level {
			levelName = l.Name
			break
		}
	}

	for _, loc := range org.Semantic.Locations {
		if loc.ID == user.Semantic.Location {
			locationName = loc.Name
			break
		}
	}

	prompt := fmt.Sprintf(`You are an expert recruiter analyzing candidates in the organization "%s".
Given the following user profile and a search query, explain why this user might or might not fit the query.

Query:
%s

User:
Name: %s
Description: %s
Team: %s
Level: %s
Location: %s

Provide a concise, analytical explanation in natural English. Focus on reasoning, not summary.`, org.Title, query, user.Name, user.Semantic.Description, teamName, levelName, locationName)

	resp, err := s.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT5Mini,
			Messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleSystem, Content: "You are a helpful AI assistant providing analytical reasoning about user fit."},
				{Role: openai.ChatMessageRoleUser, Content: prompt},
			},
		},
	)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from model")
	}

	return resp.Choices[0].Message.Content, nil
}
