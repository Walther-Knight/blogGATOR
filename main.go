package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/Walther-Knight/blogGATOR/internal/config"
	"github.com/Walther-Knight/blogGATOR/internal/database"
	"github.com/Walther-Knight/blogGATOR/internal/rss"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type state struct {
	db *database.Queries
	*config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	cmds map[string]func(*state, command) error
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	// Get a pointer string we can use for identification
	//handlerPtr := fmt.Sprintf("%p", handler)

	//fmt.Printf("Creating middleware for handler %s\n", handlerPtr)

	return func(s *state, cmd command) error {
		//fmt.Printf("Executing middleware for command '%s' with original handler %s\n",
		//	cmd.name, handlerPtr)
		user, err := s.db.GetUser(context.Background(), s.Config.CurrentUserName)
		if err != nil {
			return err
		}
		return handler(s, cmd, user)
	}
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("invalid command: username required")
	}

	userName := cmd.args[0]

	_, err := s.db.GetUser(context.Background(), userName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("user %s does not exist", userName)
		}
		return fmt.Errorf("login failed: %w", err)
	}

	err = s.SetUser(userName)
	if err != nil {
		return fmt.Errorf("set user failed: %w", err)
	}
	fmt.Println("User value set in config")
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf(("invalid command: name required"))
	}

	userName := cmd.args[0]

	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      userName,
	})
	if err != nil {
		return fmt.Errorf("create user failed: %w", err)
	}

	err = s.SetUser(cmd.args[0])
	if err != nil {
		return fmt.Errorf("failed to save user to config: %w", err)
	}
	fmt.Printf("User %s created successfully\n", userName)
	fmt.Printf("User data: %+v\n", user)
	return nil
}

func handlerReset(s *state, cmd command) error {
	if len(cmd.args) != 0 {
		return fmt.Errorf(("invalid command: no argument required"))
	}

	err := s.db.ResetFeedFollows(context.Background())
	if err != nil {
		return fmt.Errorf("error resetting user table: %w", err)
	}

	err2 := s.db.ResetUsers(context.Background())
	if err2 != nil {
		return fmt.Errorf("error resetting user table: %w", err2)
	}

	err3 := s.db.ResetFeeds(context.Background())
	if err3 != nil {
		return fmt.Errorf("error resetting user table: %w", err3)
	}

	return nil
}

func handlerGetAllUsers(s *state, cmd command) error {
	if len(cmd.args) != 0 {
		return fmt.Errorf(("invalid command: no argument required"))
	}

	userList, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("error getting users: %w", err)
	}

	loggedUser := s.Config.CurrentUserName

	for _, user := range userList {
		if user.Name == loggedUser {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}
	return nil
}

func handlerAgg(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf(("invalid command: syntax agg <timeBetweenReqs>"))
	}

	timeBetweenReqs, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return fmt.Errorf("invalid command: duration should be valid: %w", err)
	}

	fmt.Printf("Collecting feeds every %v", timeBetweenReqs)
	ticker := time.NewTicker(timeBetweenReqs)
	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}

}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) > 2 {
		return fmt.Errorf(("invalid command: too many arguments usage 'addfeed <name> <url>'"))
	}

	if len(cmd.args) < 2 {
		return fmt.Errorf("invalid command: usage 'addfeed <name> <url>'")
	}

	name := cmd.args[0]
	url := cmd.args[1]

	feed, err2 := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
		Url:       url,
		UserID:    user.ID,
	})
	if err2 != nil {
		return fmt.Errorf("error creating feed in database: %w", err2)
	}

	_, err3 := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})
	if err3 != nil {
		return fmt.Errorf("error creating feed follow: %w", err3)
	}

	fmt.Println(feed)
	return nil
}

func handlerListFeeds(s *state, cmd command) error {
	if len(cmd.args) != 0 {
		return fmt.Errorf(("invalid command: no argument required"))
	}

	feeds, err := s.db.GetAllFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("error retrieving feeds from database: %w", err)
	}
	for _, feed := range feeds {
		fmt.Printf("Feed Name: %s\n Feed URL: %s\n Username: %s\n", feed.Name, feed.Url, feed.Username.String)
	}

	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf(("invalid command: url required"))
	}

	url := cmd.args[0]

	currentFeed, err2 := s.db.GetFeed(context.Background(), url)
	if err2 != nil {
		return fmt.Errorf("error retrieving feed: %w", err2)
	}

	followRes, err3 := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    currentFeed.ID,
	})
	if err3 != nil {
		return fmt.Errorf("error creating feed follow: %w", err3)
	}

	for _, feed := range followRes {
		fmt.Printf("Feed: %s User: %s\n", feed.FeedName, feed.UserName)
	}

	return nil
}

func handlerUnFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf(("invalid command: url required"))
	}

	url := cmd.args[0]

	currentFeed, err := s.db.GetFeed(context.Background(), url)
	if err != nil {
		return fmt.Errorf("error retrieving feed: %w", err)
	}

	err2 := s.db.DeleteFollow(context.Background(), database.DeleteFollowParams{
		UserID: user.ID,
		FeedID: currentFeed.ID,
	})
	if err2 != nil {
		return fmt.Errorf("error creating feed follow: %w", err2)
	}

	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 0 {
		return fmt.Errorf(("invalid command: no arguments required"))
	}

	followingRes, err2 := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err2 != nil {
		return fmt.Errorf("error getting feed follows for user: %w", err2)
	}

	for _, feed := range followingRes {
		fmt.Println(feed.Name)
	}

	return nil
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	var limit string
	if len(cmd.args) == 0 {
		limit = "2"
	} else {
		limit = cmd.args[0]
	}
	parsedLimit, err := strconv.ParseInt(limit, 10, 32)
	if err != nil {
		return fmt.Errorf("error converting limit argument to integer: %w", err)
	}
	convLimit := int32(parsedLimit)

	browseRes, err2 := s.db.GetPostsByUser(context.Background(), database.GetPostsByUserParams{
		UserID: user.ID,
		Limit:  convLimit,
	})
	if err2 != nil {
		return fmt.Errorf("error retrieving posts by user from database: %w", err2)
	}

	for i, post := range browseRes {
		fmt.Printf("Post Number %d\n", i+1)
		fmt.Println(post.Title)
		fmt.Println(post.Description)
		fmt.Println(post.PublishedAt)
		fmt.Println(post.Url)
	}

	return nil
}

func scrapeFeeds(s *state) error {
	nextFeed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("error fetching next feed: %w", err)
	}

	err2 := s.db.MarkFeed(context.Background(), database.MarkFeedParams{
		LastFetchedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
		ID: nextFeed.ID,
	})
	if err2 != nil {
		return fmt.Errorf("error marking feed fetched: %w", err2)
	}

	currentFeed, err := rss.FetchFeed(context.Background(), nextFeed.Url)
	if err != nil {
		return fmt.Errorf("error fetching feed: %w", err)
	}

	for _, item := range currentFeed.Channel.Item {
		parsedDate, err2 := time.Parse("Mon, 2 Jan 2006 15:04:05 -0700", item.PubDate)
		if err2 != nil {
			fmt.Printf("error parsing blog time: %v\n", err2)
			continue
		}
		err3 := s.db.CreatePost(context.Background(), database.CreatePostParams{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Title:     item.Title,
			Url:       item.Link,
			Description: sql.NullString{
				String: item.Description,
				Valid:  item.Description != "",
			},
			PublishedAt: parsedDate,
			FeedID:      nextFeed.ID,
		})
		if err3 != nil {
			if pqErr, ok := err3.(*pq.Error); ok {
				if pqErr.Code == "23505" {
					continue
				}
			}
			fmt.Printf("error inserting to posts table: %v\n", err3)
		}
	}

	return nil
}

func (c *commands) run(s *state, cmd command) error {
	handler, exists := c.cmds[cmd.name]
	if !exists {
		return fmt.Errorf("unknown command: %s", cmd.name)
	}
	//fmt.Printf("Executing command '%s' with handler %p\n", cmd.name, handler)
	return handler(s, cmd)
}

func (c *commands) register(name string, handler func(*state, command) error) {
	//fmt.Printf("register(): Registering command '%s' with handler %p\n", name, handler)
	c.cmds[name] = handler
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading config: %v\n", err)
		os.Exit(1)
	}
	appState := &state{Config: &cfg}
	cmds := commands{
		make(map[string]func(*state, command) error),
	}

	db, err := sql.Open("postgres", cfg.DbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
	}
	dbQueries := database.New(db)
	appState.db = dbQueries

	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerGetAllUsers)
	cmds.register("agg", handlerAgg)
	cmds.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	cmds.register("feeds", handlerListFeeds)
	cmds.register("follow", middlewareLoggedIn(handlerFollow))
	cmds.register("following", middlewareLoggedIn(handlerFollowing))
	cmds.register("unfollow", middlewareLoggedIn(handlerUnFollow))
	cmds.register("browse", middlewareLoggedIn(handlerBrowse))

	args := os.Args
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "not enough arguments\n")
		os.Exit(1)
	}

	cmd := command{
		name: args[1],
		args: args[2:],
	}

	if err := cmds.run(appState, cmd); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

}
