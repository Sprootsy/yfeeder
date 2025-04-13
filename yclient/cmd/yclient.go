package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strings"
	"time"
)

const searchURL = "http://hn.algolia.com/api/v1/search_by_date"

type CLI struct {
	MinPoints		int
	MinComments		int
	Tags			[]string
	Age				time.Duration
}

type Article struct {
	ID			string	`json:"objectID"`
	Title		string	`json:"title"`
	URL			string	`json:"url"`
	Points		int		`json:"points"`
	NumComments int		`json:"num_comments"`
}

type Hits struct {
	Articles    []Article `json:"hits"`
}

var allowedTags = map[string]bool{"story": true, "comment": true, "poll": true,
		"pollopt": true, "show_hn": true, "ask_hn": true, "front_page": true,
		"author_": false, "story_": false}

var cli CLI

func init() {
	points := flag.Int("minPoints", 50, "Set the minimum amount of points the article must have")
	comm := flag.Int("minComments", 0, "Set the minimum amount of comments the article must have")
	tags := flag.String("tags", "front_page,story,show_hn", "A comma separated list of the tags the article must have")
	a := flag.Duration("age", 24*time.Hour, "Set the age of the article in time. Valid units are: ns, us, ms, s, m and h")
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
			errMsgs = append(errMsgs, fmt.Sprintf(tag, "is not an allowed tag"))
		}
	}
	if len(errMsgs) > 0 {
		for _, em := range errMsgs {
			fmt.Println(em)
		}
		flag.PrintDefaults()
		os.Exit(1)
	}
	
	cli = CLI{
		MinPoints: *points,
		MinComments: *comm,
		Tags: t,
		Age: *a,
	}
}
// API description
// https://hn.algolia.com/api
func main() {
	
	u := buildURL(searchURL, &cli)
	log.Println("Going to invoke:", u)	
	resp, errResp := http.DefaultClient.Get(u)
	if errResp != nil {
		log.Fatalln("An error occurred while calling search api.", errResp)
	}
	/*
	
	*/
	if resp.StatusCode >= http.StatusBadRequest {
		s := bufio.NewScanner(resp.Body)
		for s.Scan() {
			log.Println(s.Text())
		}
		log.Fatalln("Error response received")
	}
	dec := json.NewDecoder(resp.Body)
	hits := &Hits{}
	if errDecode := dec.Decode(hits); errDecode != nil {
		log.Fatalln("Cannot parse api response.", errDecode)
	}
	log.Println(hits)
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
