package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

const modelName = "gemini-2.0-flash"
const envGeminiApiKey = "GEMINI_API_KEY"
const promptSummary = "Summarize this article. Write first a short description of maximum 800 words, then the key points." 

var apiKey string
var articleUrl *string 

func printResponse(resp *genai.GenerateContentResponse) {
	msg := strings.Builder{}
	if len(resp.Candidates) < 1 {
		msg.WriteString("No reply received from model.")
	}
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				if s, ok := part.(genai.Text); ok {
					msg.WriteString(string(s))
				}
			}
		}
	}
	log.Println(msg.String())
}

func init() {
	errMsg := ""
	apiKey = os.Getenv(envGeminiApiKey)
	if len(apiKey) < 1 {
		errMsg = fmt.Sprintln("Set the API Key in the", envGeminiApiKey, "env var.")
	}
	articleUrl = flag.String("url", "", "URL to the article to summarize")
	flag.Parse()

	if len(*articleUrl) < 1 {
		errMsg = fmt.Sprintln("A URL must be provided")
	}
	_, errURL := url.Parse(*articleUrl)
	if errURL != nil {
		errMsg = fmt.Sprintln("The input URL is invalid.", errURL)
	}
	if len(errMsg) > 0 {
		fmt.Println(errMsg)
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func main() {
	ctx := context.Background()
	client, errClient := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if errClient != nil {
		log.Fatalln("Cannot create genai client.", errClient)
	}
	model := client.GenerativeModel(modelName)
	model.SetMaxOutputTokens(1000)
	model.SystemInstruction = genai.NewUserContent(genai.Text("Report only what is written in the article, do not make anything up."))
	inputPrompt := fmt.Sprintf("%s %s", promptSummary, *articleUrl)
	resp, errResp := model.GenerateContent(ctx, genai.Text(inputPrompt))
	if errResp != nil {
		log.Fatalln(errResp)
	}

	printResponse(resp)
	log.Println("Total token count:", resp.UsageMetadata.TotalTokenCount)
}
