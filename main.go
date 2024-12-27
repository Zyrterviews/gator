//nolint:err113,forbidigo,exhaustruct,wrapcheck
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/zyrterviews/gator-bootdotdev/internal/config"
	"github.com/zyrterviews/gator-bootdotdev/internal/database"
	"github.com/zyrterviews/gator-bootdotdev/internal/rss"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type state struct {
	Cfg *config.Config
	DB  *database.Queries
}

type command struct {
	name string
	args []string
}

type commands struct {
	mapping map[string]func(*state, command) error
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.mapping[name] = f
}

func (c *commands) run(s *state, cmd command) error {
	fn, ok := c.mapping[cmd.name]
	if !ok {
		return fmt.Errorf("no such handler: %q", cmd.name)
	}

	return fn(s, cmd)
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("usage login <username>")
	}

	if s == nil {
		return errors.New("state is nil")
	}

	if s.Cfg == nil {
		return errors.New("no config found")
	}

	if s.Cfg.CurrentUserName == "" {
		return errors.New(
			"no username specified. Use `register <username>` first",
		)
	}

	_, err := s.DB.GetUser(context.Background(), cmd.args[0])
	if err != nil {
		return fmt.Errorf("could not get user: %w", err)
	}

	s.Cfg.SetUser(cmd.args[0])

	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("usage register <username>")
	}

	if s == nil {
		return errors.New("state is nil")
	}

	if s.DB == nil {
		return errors.New("no DB found")
	}

	if s.Cfg == nil {
		return errors.New("no config found")
	}

	user, err := s.DB.CreateUser(context.Background(), cmd.args[0])
	if err != nil {
		return fmt.Errorf("could not create user: %w", err)
	}

	s.Cfg.SetUser(cmd.args[0])

	fmt.Printf("User %s registered\n", cmd.args[0])
	fmt.Println(user)

	return nil
}

func handlerReset(s *state, _ command) error {
	if s.DB == nil {
		return errors.New("no DB found")
	}

	err := s.DB.DeleteUsers(context.Background())
	if err != nil {
		return fmt.Errorf("could not clear users table: %w", err)
	}

	return nil
}

func handlerUsers(s *state, _ command) error {
	if s == nil {
		return errors.New("state is nil")
	}

	if s.DB == nil {
		return errors.New("no DB found")
	}

	if s.Cfg == nil {
		return errors.New("no config found")
	}

	users, err := s.DB.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("could not get users: %w", err)
	}

	for _, user := range users {
		if s.Cfg.CurrentUserName == user.Name {
			fmt.Printf("* %s (current)\n", user.Name)

			continue
		}

		fmt.Printf("* %s\n", user.Name)
	}

	return nil
}

func scrapeFeeds(s *state, tickerCh <-chan time.Time) {
	if s == nil {
		panic("state is nil")
	}

	if s.DB == nil {
		panic("no DB found")
	}

	nextFeed, err := s.DB.GetNextFeedToFetch(context.Background())
	if err != nil {
		panic(err)
	}

	err = s.DB.MarkFeedFetched(context.Background(), nextFeed.ID)
	if err != nil {
		panic(err)
	}

	feed, err := rss.FetchFeed(context.Background(), nextFeed.Url)
	if err != nil {
		panic(err)
	}

	fmt.Println("Scraping", feed.Channel.Title)

	for _, item := range feed.Channel.Item {
		publishedAt, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		opts := database.CreatePostParams{
			Url:         item.Link,
			FeedID:      nextFeed.ID,
			Title:       item.Title,
			Description: item.Description,
			PublishedAt: publishedAt,
		}

		_, err = s.DB.CreatePost(context.Background(), opts)
		if err != nil {
			var pqErr *pq.Error

			if errors.As(err, &pqErr) {
				if pqErr.Constraint != "posts_url_key" ||
					pqErr.Constraint == "posts_url_key" &&
						pqErr.Code.Name() != "unique_violation" {
					fmt.Println(err)
					os.Exit(1)
				}
			} else {
				fmt.Println(fmt.Errorf("unknown error: %w", err))
				os.Exit(1)
			}
		}

		<-tickerCh
	}
}

func handleAggregator(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("usage agg <time_between_reqs (1s, 1m, 1h)>")
	}

	timeBetweenRequests, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return err
	}

	fmt.Println("Collecting feeds every", timeBetweenRequests)

	ticker := time.NewTicker(timeBetweenRequests)

	for {
		scrapeFeeds(s, ticker.C)
	}
}

func handleAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) == 0 || len(cmd.args) < 2 {
		return errors.New("usage addfeed <name> <url>")
	}

	name := cmd.args[0]
	url := cmd.args[1]

	opts := database.CreateFeedParams{
		Name:   name,
		Url:    url,
		UserID: user.ID,
	}

	feed, err := s.DB.CreateFeed(context.Background(), opts)
	if err != nil {
		return err
	}

	ffopts := database.CreateFeedFollowParams{
		FeedID: feed.ID,
		UserID: user.ID,
	}

	if _, err := s.DB.CreateFeedFollow(context.Background(), ffopts); err != nil {
		return err
	}

	return nil
}

func handleFeeds(s *state, _ command) error {
	if s.DB == nil {
		return errors.New("no DB found")
	}

	feeds, err := s.DB.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("could not get users: %w", err)
	}

	for _, feed := range feeds {
		fmt.Printf("* %s\n  %s\n  %s\n", feed.Name, feed.Url, feed.UserName)
	}

	return nil
}

func handleFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) == 0 {
		return errors.New("usage follow <url>")
	}

	feed, err := s.DB.GetFeed(context.Background(), cmd.args[0])
	if err != nil {
		return err
	}

	opts := database.CreateFeedFollowParams{
		FeedID: feed.ID,
		UserID: user.ID,
	}

	feedFollow, err := s.DB.CreateFeedFollow(context.Background(), opts)
	if err != nil {
		return err
	}

	fmt.Printf("%s: %s\n", feedFollow.FeedName, feedFollow.UserName)

	return nil
}

func handleFollowing(s *state, _ command, user database.User) error {
	feedFollows, err := s.DB.GetFeedFollowsForUser(
		context.Background(),
		user.Name,
	)
	if err != nil {
		return err
	}

	fmt.Printf("%s is following:\n", user.Name)

	for _, ff := range feedFollows {
		fmt.Printf("\t* %s\n", ff.FeedName)
	}

	return nil
}

func handleUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) == 0 {
		return errors.New("usage addfeed <name> <url>")
	}

	opts := database.DeleteFeedFollowParams{
		Name: user.Name,
		Url:  cmd.args[0],
	}

	if err := s.DB.DeleteFeedFollow(context.Background(), opts); err != nil {
		return err
	}

	fmt.Println("Follow removed")

	return nil
}

func handleBrowse(s *state, cmd command, user database.User) error {
	limit := 2

	if len(cmd.args) > 0 {
		var err error

		limit, err = strconv.Atoi(cmd.args[0])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	opts := database.GetPostsParams{
		Name:  user.Name,
		Limit: int32(limit),
	}

	posts, err := s.DB.GetPosts(context.Background(), opts)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, post := range posts {
		fmt.Println(post.Title)
	}

	return nil
}

func handlerHelp(_ *state, _ command) error {
	fmt.Println("gator - usage:")
	fmt.Println("help - display this message")
	fmt.Println("login <username> - logs in to the database")

	return nil
}

func middlewareLoggedIn(
	handler func(s *state, cmd command, user database.User) error,
) func(*state, command) error {
	return func(s *state, cmd command) error {
		if s == nil {
			return errors.New("state is nil")
		}

		if s.DB == nil {
			return errors.New("no DB found")
		}

		if s.Cfg == nil {
			return errors.New("no config found")
		}

		if s.Cfg.CurrentUserName == "" {
			return errors.New(
				"no username specified. Use `register <username>` first",
			)
		}

		user, err := s.DB.GetUser(context.Background(), s.Cfg.CurrentUserName)
		if err != nil {
			return err
		}

		return handler(s, cmd, user)
	}
}

func main() {
	args := os.Args
	if len(args) == 1 {
		_ = handlerHelp(nil, command{})

		os.Exit(1)
	}

	cfg := config.Read()

	db, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	dbQueries := database.New(db)
	env := &state{Cfg: &cfg, DB: dbQueries}
	cmds := commands{mapping: make(map[string]func(*state, command) error)}

	cmds.register("help", handlerHelp)
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerUsers)
	cmds.register("agg", handleAggregator)
	cmds.register("addfeed", middlewareLoggedIn(handleAddFeed))
	cmds.register("feeds", handleFeeds)
	cmds.register("follow", middlewareLoggedIn(handleFollow))
	cmds.register("following", middlewareLoggedIn(handleFollowing))
	cmds.register("unfollow", middlewareLoggedIn(handleUnfollow))
	cmds.register("browse", middlewareLoggedIn(handleBrowse))

	cmd := command{
		name: args[1],
		args: args[2:],
	}

	if err := cmds.run(env, cmd); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
