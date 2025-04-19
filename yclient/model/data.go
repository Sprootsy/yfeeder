package model

type Article struct {
	ID          string `json:"objectID"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	Points      int    `json:"points"`
	NumComments int    `json:"num_comments"`
}

type Hits struct {
	Articles []Article `json:"hits"`
}
