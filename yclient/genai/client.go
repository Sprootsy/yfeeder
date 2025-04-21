package genai

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

const modelName = "gemini-2.0-flash"
const promptSummary = "Summarize this article. Write first a short description of maximum 800 words, then the key points."

type Client struct {
	model   *genai.GenerativeModel
}

// NewClient creates a new genai client
func NewClient(apiKey string) *Client {
	ctx := context.Background()
	client, errClient := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if errClient != nil {
			log.Fatalln("Cannot create genai client.", errClient)
	}
	model := client.GenerativeModel(modelName)
	model.SetMaxOutputTokens(1000)
	model.SystemInstruction = genai.NewUserContent(genai.Text("Report only what is written in the article, do not make anything up. Print the answer in valid HTML, inside a <div> element. Don't include any conversation in the answer."))
	return &Client{
		model: model,
	}
}

// GenerateSummaryFromURL Asks the genAI model to make a summary of the given URl
func (c Client) GenerateSummaryFromURL(u string) (*genai.GenerateContentResponse, error) {
	log.Println("Generating summary for url:", u)
	if _, errUrl := url.Parse(u); errUrl != nil {
		return nil, errUrl
	}
	inputPrompt := fmt.Sprintf("%s %s", promptSummary, u)
	ctx, cancFunc := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancFunc()
	log.Println("Done generating summary for url:", u)
	return c.model.GenerateContent(ctx, genai.Text(inputPrompt))
}

// ReadResponse reads the reply from the model
func ReadResponse(resp *genai.GenerateContentResponse) string {
	log.Println("Reading response")
	msg := strings.Builder{}
	if len(resp.Candidates) < 1 {
			msg.WriteString("No reply received from model.")
	}
	for _, cand := range resp.Candidates {
			if cand.Content != nil {
					for _, part := range cand.Content.Parts {
							if s, ok := part.(genai.Text); ok {
									msg.WriteString(string(s))
							} else {
									// for now just log if something else is returned
									log.Println("Received a genai.Part which is not text; skipping.")
							}
					}
			}
	}
	log.Println("Done reading response")
	res, _ := strings.CutSuffix(msg.String(), "```")
	res, _ = strings.CutPrefix(res, "```html")
	return res
}
