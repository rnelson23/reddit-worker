package main

import (
	"encoding/json"
	"github.com/JamesPEarly/loggly"
	"github.com/jzelinskie/geddit"
	"sort"
	"time"
)

type Data struct {
	Posts []Post
}

type Post struct {
	FullID  string
	Author  string
	Created float64
	Title   string
	URL     string
}

func main() {
	logger := loggly.New("worker")

	session, _ := geddit.NewOAuthSession(
		"sCjeFwY1yd98IFwM-VaIeQ",
		"6LnIpUYTYePRJoay8RPn0qAhlz25BA",
		"gedditAgent v1",
		"http://redirect.url",
	)

	_ = session.LoginAuth("sen-senpai", "Squeek247")
	_ = logger.EchoSend("info", "Ready!")

	options := geddit.ListingOptions{Limit: 5, Before: "t3_x1zcfh"}

	for range time.Tick(time.Second * 2) {
		submissions, _ := session.SubredditSubmissions("FloridaMan", geddit.NewSubmissions, options)

		if len(submissions) == 0 {
			continue
		}

		var data Data

		for _, s := range submissions {
			post := Post{
				FullID:  s.FullID,
				Author:  s.Author,
				Created: s.DateCreated,
				Title:   s.Title,
				URL:     s.URL,
			}

			data.Posts = append(data.Posts, post)
		}

		sort.Slice(data.Posts, func(i, j int) bool {
			return data.Posts[i].Created < data.Posts[j].Created
		})

		options = geddit.ListingOptions{Limit: 5, Before: data.Posts[len(data.Posts)-1].FullID}
		bytes, _ := json.MarshalIndent(data, "", "    ")

		_ = logger.EchoSend("info", string(bytes))
	}
}
