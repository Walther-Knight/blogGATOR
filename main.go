package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/Walther-Knight/blogGATOR/internal/config"
	"github.com/Walther-Knight/blogGATOR/internal/database"
	"github.com/Walther-Knight/blogGATOR/internal/rss"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
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

func handlerResetUsers(s *state, cmd command) error {
	if len(cmd.args) != 0 {
		return fmt.Errorf(("invalid command: no argument required"))
	}

	err := s.db.ResetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("error resetting user table: %w", err)
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
	if len(cmd.args) != 0 {
		return fmt.Errorf(("invalid command: no argument required"))
	}

	urlDefault := "https://www.wagslane.dev/index.xml"
	testFeed, err := rss.FetchFeed(context.Background(), urlDefault)
	if err != nil {
		return fmt.Errorf("error fetching feed: %w", err)
	}
	fmt.Println(testFeed)

	return nil
}

func (c *commands) run(s *state, cmd command) error {
	handler, exists := c.cmds[cmd.name]
	if !exists {
		return fmt.Errorf("unknown command: %s", cmd.name)
	}
	return handler(s, cmd)
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.cmds[name] = f
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
	cmds.register("reset", handlerResetUsers)
	cmds.register("users", handlerGetAllUsers)
	cmds.register("agg", handlerAgg)

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
