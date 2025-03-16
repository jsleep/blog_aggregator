package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"internal/config"
	"io"
	"net/http"
	"os"
	"strconv"
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

	user, err := s.db.GetUserByName(context.Background(), name)
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

func addFeedHandler(s *state, cmd command, user database.User) error {
	// Check if the command is "register"
	if cmd.Command != "addfeed" {
		return fmt.Errorf("invalid command")
	}

	// Check if the arguments are valid
	if len(cmd.Args) < 2 {
		return fmt.Errorf("missing name/url arguments")
	}

	name := cmd.Args[0]
	url := cmd.Args[1]

	// add feed ti database
	feed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        uuid.New(),
		Name:      name,
		Url:       url,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
	})

	if err != nil {
		return err
	}

	// current user should follow feed they just added
	_, err = s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		UserID:    user.ID,
		FeedID:    feed.ID,
		CreatedAt: time.Now(),
	})
	if err != nil {
		return err
	}

	fmt.Printf("Feed %s added successfully\n", feed)

	return nil
}

func followHandler(s *state, cmd command, user database.User) error {
	// Check if the command is "register"
	if cmd.Command != "follow" {
		return fmt.Errorf("invalid command")
	}

	// Check if the arguments are valid
	if len(cmd.Args) < 1 {
		return fmt.Errorf("missing url arg")
	}

	url := cmd.Args[0]

	feed, err := s.db.GetFeed(context.Background(), url)
	if err != nil {
		return err
	}

	// get current user from db by name
	_, err = s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		UserID:    user.ID,
		FeedID:    feed.ID,
		CreatedAt: time.Now(),
	})

	if err != nil {
		return err
	}

	fmt.Printf(" User %s followed feed %s successfully\n", user.Name, feed.Name)

	return nil
}

func unfollowHandler(s *state, cmd command, user database.User) error {
	// Check if the command is "register"
	if cmd.Command != "unfollow" {
		return fmt.Errorf("invalid command")
	}

	// Check if the arguments are valid
	if len(cmd.Args) < 1 {
		return fmt.Errorf("missing url arg")
	}

	url := cmd.Args[0]

	feed, err := s.db.GetFeed(context.Background(), url)
	if err != nil {
		return err
	}

	// remove feedfollow
	err = s.db.RemoveFeedFollow(context.Background(), database.RemoveFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return err
	}

	fmt.Printf(" User %s unfollowed feed %s successfully\n", user.Name, feed.Name)

	return nil
}

func followingHandler(s *state, cmd command, user database.User) error {
	// Check if the command is "register"
	if cmd.Command != "following" {
		return fmt.Errorf("invalid command")
	}

	// get current user from db by name
	follows, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)

	if err != nil {
		return err
	}

	fmt.Printf("User %s follows:\n", user.Name)

	for _, follow := range follows {
		fmt.Printf("* %s\n", follow.FeedName)
	}
	return nil
}

func listFeedsHandler(s *state, cmd command) error {
	// Check if the command is "register"
	if cmd.Command != "feeds" {
		return fmt.Errorf("invalid command")
	}

	// get current user from db by name
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return err
	}

	for _, feed := range feeds {
		// get feed from db by name
		fmt.Printf("* %s, ", feed.Name)
		fmt.Printf("  %s, ", feed.Url)

		user, err := s.db.GetUserById(context.Background(), feed.UserID)
		if err != nil {
			return err
		}
		fmt.Printf("%s \n", user.Name)
	}

	return nil
}

func scrapeFeeds(s *state) error {
	next_feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		panic(err)
	}

	s.db.MarkFeedFetched(context.Background(), database.MarkFeedFetchedParams{
		ID:            next_feed.ID,
		LastFetchedAt: sql.NullTime{Time: time.Now(), Valid: true},
	})

	fmt.Printf("Fetched feed %s at %s\n", next_feed.Name, time.Now().Format(time.RFC3339))

	feed, err := fetchFeed(context.Background(), next_feed.Url)
	if err != nil {
		panic(err)
	}

	// Print entire feed struct
	for _, item := range feed.Channel.Item {
		fmt.Printf("* Item: %s", item.Title)
		fmt.Printf(", Time: %s\n", item.PubDate)
		parseTime, err := parseTime(item.PubDate)
		if err != nil {
			fmt.Printf("Error parsing time: %v\n", err)
			continue
		}

		post, err := s.db.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			FeedID:      next_feed.ID,
			Title:       item.Title,
			Url:         item.Link,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Description: item.Description,
			PublishedAt: parseTime,
		})

		if err != nil {
			fmt.Printf("Error creating post: %v\n", err)
			continue
		} else {
			fmt.Printf("Post created: %s %s\n", post.Title, post.Url)
		}
	}

	return nil
}

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	r.Header.Set("User-Agent", "gator")
	resp, err := client.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch feed: %s", resp.Status)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var feed RSSFeed

	if err := xml.Unmarshal(b, &feed); err != nil {
		return nil, err
	}

	// unescape html strings
	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)
	for i := range feed.Channel.Item {
		feed.Channel.Item[i].Title = html.UnescapeString(feed.Channel.Item[i].Title)
		feed.Channel.Item[i].Description = html.UnescapeString(feed.Channel.Item[i].Description)
	}

	return &feed, nil
}

func aggregationHandler(s *state, cmd command) error {
	// Check if the command is "register"
	if cmd.Command != "agg" {
		return fmt.Errorf("invalid command")
	}

	// Check if the arguments are valid
	if len(cmd.Args) < 1 {
		return fmt.Errorf("missing duration arguments")
	}

	// name := cmd.Args[0]
	time_between_reqs, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("invalid duration: %v", err)
	}

	ticker := time.NewTicker(time_between_reqs)
	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
}

func browseHandler(s *state, cmd command, user database.User) error {
	// Check if the command is "register"
	if cmd.Command != "browse" {
		return fmt.Errorf("invalid command")
	}

	var limit int = 2
	var err error

	// Check if the arguments are valid
	if len(cmd.Args) >= 1 {
		limit, err = strconv.Atoi(cmd.Args[0])
		if err != nil {
			return err
		}
	}

	// name := cmd.Args[0]

	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{UserID: user.ID, Limit: int32(limit)})
	if err != nil {
		return err
	}

	for _, post := range posts {
		fmt.Printf("* %s, ", post.Title)
		fmt.Printf("  %s, ", post.Url)
		fmt.Printf("%s \n", post.PublishedAt.Format(time.RFC3339))
	}
	return nil
}

func parseTime(timeStr string) (time.Time, error) {
	// Try a few common formats
	formats := []string{
		time.RFC1123Z,
		time.RFC1123,
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"Mon, 02 Jan 2006 15:04:05 -0700",
		// Add more formats as needed
	}

	var t time.Time
	var err error

	for _, format := range formats {
		t, err = time.Parse(format, timeStr)
		if err == nil {
			return t, nil
		}
	}

	// If we get here, none of our formats worked
	return time.Time{}, fmt.Errorf("could not parse time: %s", timeStr)
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {

		user, err := s.db.GetUserByName(context.Background(), s.Config.User)
		if err != nil {
			return err
		}

		return handler(s, cmd, user)
	}
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
	commands.register("agg", aggregationHandler)
	commands.register("addfeed", middlewareLoggedIn(addFeedHandler))
	commands.register("feeds", listFeedsHandler)
	commands.register("follow", middlewareLoggedIn(followHandler))
	commands.register("following", middlewareLoggedIn(followingHandler))
	commands.register("unfollow", middlewareLoggedIn(unfollowHandler))
	commands.register("browse", middlewareLoggedIn(browseHandler))

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
