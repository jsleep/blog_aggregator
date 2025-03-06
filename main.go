package main

import (
	"fmt"
	"internal/config"
	"os"
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

func handlerLogin(s *state, cmd command) error {
	// Check if the command is "login"
	if cmd.Command != "login" {
		return fmt.Errorf("invalid command")
	}

	// Check if the arguments are valid
	if len(cmd.Args) < 1 {
		return fmt.Errorf("missing arguments")
	}

	// Set the user in the configuration
	err := s.Config.SetUser(cmd.Args[0])
	if err != nil {
		return err
	}

	fmt.Printf("User set to %s\n", cmd.Args[0])

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

	commands := make(commands)
	commands.register("login", handlerLogin)

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
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	fmt.Println("Command executed successfully")

}
