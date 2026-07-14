package db

import (
	"context"
	"fmt"
	"log"

	"math/rand"

	"github.com/alimohammadi/golan-social.git/internal/store"
)

var usernames = []string{
	"alice", "bob", "charlie", "diana", "eve", "frank", "grace", "henry",
	"iris", "jack", "kate", "liam", "mia", "noah", "olivia", "peter",
	"quinn", "rachel", "sam", "tina", "uma", "victor", "wendy", "xander",
	"yara", "zoe",
}

var titles = []string{
	"10 Tips for Better Code", "Understanding Concurrency in Go", "Building REST APIs",
	"Introduction to Docker", "Kubernetes for Beginners", "Clean Architecture Explained",
	"Why Go is Great", "Mastering SQL Joins", "Microservices vs Monolith",
	"Getting Started with Redis", "How to Write Unit Tests", "DevOps Best Practices",
	"GraphQL vs REST", "CI/CD Pipeline Setup", "Securing Your API",
	"Optimizing Database Queries", "Event-Driven Architecture", "Working with WebSockets",
	"Caching Strategies", "Building Scalable Systems",
}

var tags = []string{
	"go", "backend", "api", "database", "docker", "kubernetes", "testing",
	"performance", "security", "architecture", "microservices", "devops",
	"sql", "nosql", "redis", "caching", "concurrency", "cloud", "grpc", "rest",
}

var comments = []string{
	"Great post, very helpful!", "Thanks for sharing this.", "I learned a lot from this.",
	"Could you elaborate on that point?", "This is exactly what I needed.",
	"Well explained, keep it up!", "Interesting perspective.", "I disagree with this approach.",
	"Have you tried using channels instead?", "This saved me hours of debugging.",
	"Love the examples provided.", "Looking forward to the next post.",
	"Can you share the source code?", "This works perfectly, thank you!",
	"Very well written article.", "I had the same issue, this solved it.",
}

func Seed(storage store.Storage) {
	ctx := context.Background()

	users := generateUsers(100)
	for _, user := range users {
		if err := storage.Users.Create(ctx, nil, user); err != nil {
			log.Println("Error creating user: ", err)
			return
		}
	}

	posts := generatePosts(200, users)
	for _, post := range posts {
		if err := storage.Posts.Create(ctx, post); err != nil {
			log.Println("Error creating post: ", err)
			return
		}
	}

	cms := generateComments(500, users, posts)
	for _, comment := range cms {
		if err := storage.Comments.Create(ctx, comment); err != nil {
			log.Println("Error creating comment: ", err)
			return
		}
	}

	log.Println("Seeding complete")
}

func generateUsers(num int) []*store.User {
	users := make([]*store.User, num)
	// role := getRoleByName("user")

	for i := range num {
		users[i] = &store.User{
			Username: usernames[i%len(usernames)] + fmt.Sprintf("%d", i),
			Email:    usernames[i%len(usernames)] + fmt.Sprintf("%d", i) + "@gmail.com",
			Role: store.Role{
				Name: "user",
			},
		}
		_ = users[i].Password.Set("123123")
	}

	return users
}

func generatePosts(num int, users []*store.User) []*store.Post {
	posts := make([]*store.Post, num)

	for i := range num {
		user := users[rand.Intn(len(users))]

		posts[i] = &store.Post{
			UserID:  user.ID,
			Title:   titles[rand.Intn(len(titles))],
			Content: titles[rand.Intn(len(titles))],
			Tags: []string{
				tags[rand.Intn(len(tags))],
				tags[rand.Intn(len(tags))],
			},
		}
	}

	return posts
}

func generateComments(num int, users []*store.User, posts []*store.Post) []*store.Comment {
	cms := make([]*store.Comment, num)
	for i := range num {
		cms[i] = &store.Comment{
			PostID:  posts[rand.Intn(len(posts))].ID,
			UserID:  users[rand.Intn(len(users))].ID,
			Content: comments[rand.Intn(len(comments))],
		}
	}

	return cms
}
