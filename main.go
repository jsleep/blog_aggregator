package main

import (
	"context"
	"database/sql"
	"fmt"
	"internal/config"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jsleep/blog_aggregator/internal/database"
	_ "github.com/lib/pq"
)

type command struct {
	// Command is the command to execute
	Command string
	// Args is the arguments to the command
	Args []string
}

type state struct {
	// Config is the configuration
	Config *config.Config

	// db is the database connection
	db *database.Queries
}

type commands map[string]func(*state, command) error

func (c *commands) register(name string, f func(*state, command) error) {
	(*c)[name] = f
}

func (c *commands) run(s *state, cmd command) error {
	handler, ok := (*c)[cmd.Command]
	if !ok {
		return fmt.Errorf("command %s not found", cmd.Command)
	}
	return handler(s, cmd)
}

func loginHandler(s *state, cmd command) error {
	// Check if the command is "login"
	if cmd.Command != "login" {
		return fmt.Errorf("invalid command")
	}

	// Check if the arguments are valid
	if len(cmd.Args) < 1 {
		return fmt.Errorf("missing name argument")
	}

	name := cmd.Args[0]

	user, err := s.db.GetUser(context.Background(), name)
	if err != nil {
		return err
	}

	// Set the user in the configuration
	err = s.Config.SetUser(user.Name)
	if err != nil {
		return err
	}

	fmt.Printf("User set to %s\n", user.Name)

	return nil
}

func resetHandler(s *state, cmd command) error {
	// Check if the command is "login"
	if cmd.Command != "reset" {
		return fmt.Errorf("invalid command")
	}

	err := s.db.DeleteUsers(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func listUsersHandler(s *state, cmd command) error {
	// Check if the command is "login"
	if cmd.Command != "users" {
		return fmt.Errorf("invalid command")
	}

	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}

	for _, user := range users {
		current := ""
		if user.Name == s.Config.User {
			current = " (current)"
		}
		fmt.Printf("* %s%s\n", user.Name, current)
	}

	return nil
}

func registerHandler(s *state, cmd command) error {
	// Check if the command is "register"
	if cmd.Command != "register" {
		return fmt.Errorf("invalid command")
	}

	// Check if the arguments are valid
	if len(cmd.Args) < 1 {
		return fmt.Errorf("missing name arguments")
	}

	name := cmd.Args[0]

	// Register the user in the database
	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	if err != nil {
		return err
	}

	fmt.Printf("User %s registered successfully\n", user.Name)

	err = s.Config.SetUser(user.Name)
	if err != nil {
		return err
	}

	fmt.Printf("User data  %s\n", user)

	return nil
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		panic(err)
	}

	state := &state{
		Config: &cfg,
	}

	db, err := sql.Open("postgres", cfg.DBUrl)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	dbQueries := database.New(db)
	state.db = dbQueries

	commands := make(commands)
	commands.register("login", loginHandler)
	commands.register("register", registerHandler)
	commands.register("reset", resetHandler)
	commands.register("users", listUsersHandler)

	args := os.Args
	if len(args) < 2 {
		fmt.Println("Usage: gator <command> [args]")
		os.Exit(1)
	}

	cmd := command{
		Command: args[1],
		Args:    args[2:],
	}

	err = commands.run(state, cmd)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Command executed successfully")

}
