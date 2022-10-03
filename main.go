package main

import (
	"encoding/json"
	"github.com/JamesPEarly/loggly"
	"github.com/jzelinskie/geddit"
	"time"
)

type Data struct {
	Posts [5]Post
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
			break
		}

		var data Data

		for i, s := range submissions {
			post := Post{
				FullID:  s.FullID,
				Author:  s.Author,
				Created: s.DateCreated,
				Title:   s.Title,
				URL:     s.URL,
			}

			data.Posts[4-i] = post
		}

		options = geddit.ListingOptions{Limit: 5, Before: data.Posts[2].FullID}
		bytes, _ := json.MarshalIndent(data, "", "    ")

		_ = logger.EchoSend("info", string(bytes))
	}
}
