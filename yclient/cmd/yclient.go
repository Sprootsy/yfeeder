package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
	"yclient/genai"
	"yclient/model"
	"yclient/templates"
)

const searchURL = "http://hn.algolia.com/api/v1/search_by_date"

type CLI struct {
	MinPoints   int
	MinComments int
	Tags        []string
	Age         time.Duration
	OutDir      string
}

var allowedTags = map[string]bool{
	"story": true, "comment": true, "poll": true,
	"pollopt": true, "show_hn": true, "ask_hn": true, "front_page": true,
	"author_": false, "story_": false,
}

const (
	articlesTemplate = "articles.html"
	envGenaiApiKey   = "GENAI_API_KEY"
)

var (
	apiKey string
	cli    CLI
)

// API description
// https://hn.algolia.com/api
func main() {
	parseFlags()
	u := buildURL(searchURL, &cli)
	log.Println("Going to invoke:", u)
	resp, errResp := http.DefaultClient.Get(u)
	if errResp != nil {
		log.Fatalln("An error occurred while calling search api.", errResp)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		s := bufio.NewScanner(resp.Body)
		for s.Scan() {
			log.Println(s.Text())
		}
		log.Fatalln("Error response received.")
	}
	log.Println("Respones status code:", resp.StatusCode)
	dec := json.NewDecoder(resp.Body)
	hits := &model.Hits{}
	if errDecode := dec.Decode(hits); errDecode != nil {
		log.Fatalln("Cannot parse api response.", errDecode)
	}
	log.Println(hits)

	log.Println("Going to render articles list")
	ctx, cancelFun := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelFun()
	articlesRendering := templates.ArticlesRendering{
		Rendering: templates.Rendering{
			Name:     "articles.html",
			Path:     cli.OutDir,
			Template: templates.GetTemplate(templates.Files[0]),
		},
		Data: hits.Articles,
	}
	if errAR := articlesRendering.Render(); errAR != nil {
		log.Println("Cannot render articles. Error:", errAR)
	}
	ctx.Done()

	aic := genai.NewClient(apiKey)
	for _, a := range hits.Articles {
		res, errAIS := aic.GenerateSummaryFromURL(a.URL)
		if errAIS != nil {
			log.Println("Cannot get a summary for article. Title:", a.Title,
				"Error:", errAIS)
		}
		a.Summary = genai.ReadResponse(res)

		ctx, cancelFun = context.WithTimeout(context.Background(), 2*time.Second)
		defer cancelFun()
		sumRendering := templates.SummariesRendering{
			Rendering: templates.Rendering{
				Name:     "Summaries",
				Path:     path.Join(cli.OutDir, "summaries"),
				Template: templates.GetTemplate(templates.Files[1]),
			},
			Data: a,
		}
		if errSR := sumRendering.Render(); errSR != nil {
			log.Println("Cannot render summaries. Error:", errSR)
		}
		ctx.Done()
		log.Println("Pausing for rate limiting")
		time.Sleep(5*time.Second)
	}
	log.Println("Done with everything!")
}

func parseFlags() {
	points := flag.Int("minPoints", 50, "Set the minimum amount of points the article must have")
	comm := flag.Int("minComments", 0, "Set the minimum amount of comments the article must have")
	tags := flag.String("tags", "front_page,story,show_hn", "A comma separated list of the tags the article must have")
	a := flag.Duration("age", 24*time.Hour, "Set the age of the article in time. Valid units are: ns, us, ms, s, m and h")
	outDir := flag.String("outDir", ".", "Path to the directory where the output must be written")
	ak := flag.String("apiKeyFile", "", "Path to a file where the gen ai api key can be found")
	flag.Parse()

	errMsgs := []string{}
	if *points < 0 {
		errMsgs = append(errMsgs, "minPoints must be >= 0.")
	}
	if *comm < 0 {
		errMsgs = append(errMsgs, "minComments must be >= 0.")
	}
	t := strings.Split(*tags, ",")
	for _, tag := range t {
		if allowed, found := allowedTags[tag]; !(allowed && found) {
			errMsgs = append(errMsgs, fmt.Sprintf(tag, "is not an allowed tag."))
		}
	}
	// output dir
	o := "."
	if len(*outDir) > 0 {
		o = *outDir
	}
	fi, errStat := os.Stat(*outDir)
	if os.IsNotExist(errStat) {
		errMsgs = append(errMsgs, fmt.Sprint("Directory does not exist: ", *outDir))
	}
	if errStat == nil && !fi.IsDir() {
		errMsgs = append(errMsgs, fmt.Sprint(*outDir, "must be a directory."))
	}
	// api key
	if len(*ak) > 0 {
		if fd, errOpenAK := os.Open(*ak); errOpenAK != nil {
			errMsgs = append(errMsgs, fmt.Sprint("Cannot open api key file. Error: ", errOpenAK))
		} else {
			if b, errReadAK := io.ReadAll(fd); errReadAK != nil {
				errMsgs = append(errMsgs, fmt.Sprint("Cannot read api key file. Error: ", errReadAK))
			} else {
				apiKey = string(b)
			}
			fd.Close()
		}
	} else if ak := os.Getenv(envGenaiApiKey); len(ak) > 0 {
		apiKey = ak
	} else {
		errMsgs = append(errMsgs, "Provide the genai API key either via file or", envGenaiApiKey, "env var.")
	}
	if len(errMsgs) > 0 {
		for _, em := range errMsgs {
			fmt.Println(em)
		}
		flag.PrintDefaults()
		os.Exit(1)
	}
	cli = CLI{
		MinPoints:   *points,
		MinComments: *comm,
		Tags:        t,
		Age:         *a,
		OutDir:      o,
	}
}

func buildURL(base string, cli *CLI) string {
	u := searchURL
	qp := []string{}
	numFilters := []string{}
	if len(cli.Tags) > 0 {
		qp = append(qp, fmt.Sprintf("tags=(%s)", strings.Join(cli.Tags, ",")))
	}
	if cli.MinPoints > 0 {
		numFilters = append(numFilters, fmt.Sprintf("points>=%d", cli.MinPoints))
	}
	if cli.MinComments > 0 {
		numFilters = append(numFilters, fmt.Sprintf("num_comments>=%d", cli.MinComments))
	}
	if cli.Age > 0*time.Second {
		numFilters = append(numFilters,
			fmt.Sprintf("created_at_i>=%d", int64(math.Ceil(cli.Age.Seconds()))))
	}
	qp = append(qp, fmt.Sprintf("numericFilters=%s", strings.Join(numFilters, ",")))
	return fmt.Sprintf("%s?%s", u, strings.Join(qp, "&"))
}
