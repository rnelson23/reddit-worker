package main

import (
	"context"
	"encoding/json"
	"github.com/JamesPEarly/loggly"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/joho/godotenv"
	"github.com/jzelinskie/geddit"
	"os"
	"sort"
	"time"
)

type Data struct {
	Posts []Post
}

type Post struct {
	Title       string  `dynamodbav:"title"`
	FullID      string  `dynamodbav:"name"`
	Author      string  `dynamodbav:"author"`
	Permalink   string  `dynamodbav:"permalink"`
	URL         string  `dynamodbav:"url"`
	DateCreated float64 `dynamodbav:"created_utc"`
}

func main() {
	_ = godotenv.Load()

	cfg, _ := config.LoadDefaultConfig(context.TODO())
	db := dynamodb.NewFromConfig(cfg)

	logger := loggly.New("reddit-worker")
	client, _ := geddit.NewOAuthSession(
		os.Getenv("REDDIT_CLIENT_ID"),
		os.Getenv("REDDIT_CLIENT_SECRET"),
		"gedditAgent v1",
		"http://redirect.url",
	)

	_ = client.LoginAuth(os.Getenv("REDDIT_USERNAME"), os.Getenv("REDDIT_PASSWORD"))
	_ = logger.EchoSend("info", "Ready!")

	options := geddit.ListingOptions{Before: getLastPost(db)}

	for range time.Tick(time.Hour * 1) {
		data := getPosts(client, db, options)

		if len(data.Posts) > 0 {
			options = geddit.ListingOptions{Before: data.Posts[len(data.Posts)-1].FullID}
			bytes, _ := json.MarshalIndent(data, "", "    ")

			_ = logger.EchoSend("info", string(bytes))
		}
	}
}

func getLastPost(db *dynamodb.Client) string {
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

func getPosts(client *geddit.OAuthSession, db *dynamodb.Client, options geddit.ListingOptions) Data {
	posts, _ := client.SubredditSubmissions("FloridaMan", geddit.NewSubmissions, options)
	var data Data

	for _, p := range posts {
		post := Post{
			Title:       p.Title,
			FullID:      p.FullID,
			Author:      p.Author,
			Permalink:   p.Permalink,
			URL:         p.URL,
			DateCreated: p.DateCreated,
		}

		data.Posts = append(data.Posts, post)

		item, _ := attributevalue.MarshalMap(post)
		_, _ = db.PutItem(context.TODO(), &dynamodb.PutItemInput{
			TableName: aws.String("rnelson3-reddit"),
			Item:      item,
		})
	}

	sort.Slice(data.Posts, func(i, j int) bool {
		return data.Posts[i].DateCreated < data.Posts[j].DateCreated
	})

	return data
}
