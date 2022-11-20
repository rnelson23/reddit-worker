package main

import (
	"context"
	"encoding/json"
	"github.com/JamesPEarly/loggly"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/jzelinskie/geddit"
	"os"
	"sort"
	"time"
)

type Post struct {
	Title       string  `dynamodbav:"title"`
	FullID      string  `dynamodbav:"name"`
	Author      string  `dynamodbav:"author"`
	Permalink   string  `dynamodbav:"permalink"`
	URL         string  `dynamodbav:"url"`
	DateCreated float64 `dynamodbav:"created_utc"`
}

var db *dynamodb.Client
var logger *loggly.ClientType
var client *geddit.OAuthSession

func init() {
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	db = dynamodb.NewFromConfig(cfg)

	client, _ = geddit.NewOAuthSession(
		os.Getenv("REDDIT_CLIENT_ID"),
		os.Getenv("REDDIT_CLIENT_SECRET"),
		"gedditAgent v1",
		"http://redirect.url",
	)

	_ = client.LoginAuth(os.Getenv("REDDIT_USERNAME"), os.Getenv("REDDIT_PASSWORD"))

	logger = loggly.New("rnelson3-agent")
	_ = logger.EchoSend("info", "Ready!")
}

func main() {
	options := geddit.ListingOptions{Before: getLastPost()}

	for range time.Tick(time.Hour * 1) {
		posts := getPosts(options)

		if len(posts) > 0 {
			options = geddit.ListingOptions{Before: posts[len(posts)-1].FullID}
			bytes, _ := json.MarshalIndent(posts, "", "    ")

			_ = logger.EchoSend("info", string(bytes))
		}
	}
}

func getLastPost() string {
	posts, _ := db.Scan(context.TODO(), &dynamodb.ScanInput{TableName: aws.String("rnelson3-reddit")})
	var lastPost Post

	for _, p := range posts.Items {
		var post Post
		_ = attributevalue.UnmarshalMap(p, &post)

		if post.DateCreated > lastPost.DateCreated {
			lastPost = post
		}
	}

	return lastPost.FullID
}

func getPosts(options geddit.ListingOptions) []Post {
	submissions, _ := client.SubredditSubmissions("FloridaMan", geddit.NewSubmissions, options)
	var posts []Post

	for _, s := range submissions {
		post := Post{
			Title:       s.Title,
			FullID:      s.FullID,
			Author:      s.Author,
			Permalink:   s.Permalink,
			URL:         s.URL,
			DateCreated: s.DateCreated,
		}

		posts = append(posts, post)

		item, _ := attributevalue.MarshalMap(post)
		_, _ = db.PutItem(context.TODO(), &dynamodb.PutItemInput{
			TableName: aws.String("rnelson3-reddit"),
			Item:      item,
		})
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].DateCreated < posts[j].DateCreated
	})

	return posts
}
