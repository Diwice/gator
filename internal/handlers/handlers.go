package handlers

import (
	"os"
	"fmt"
	"log"
	"time"
	"context"
	"gator/internal/rss"
	"gator/internal/state"
	"gator/internal/database"

	"github.com/google/uuid"
)

type Command struct {
	name string
	args []string
}

type Commands struct {
	cmd map[string]func(*state.State, Command) error
}

func (c *Commands) run(s *state.State, cmd Command) error {
	get_cmd := c.cmd[cmd.name]

	if err := get_cmd(s, cmd); err != nil {
		return err
	}

	return nil
}

func (c *Commands) register(name string, f func(*state.State, Command) error) {
	if c.cmd == nil {
		c.cmd = make(map[string]func(*state.State, Command) error)
	}

	c.cmd[name] = f
}

func handlerLogins(s *state.State, cmd Command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("Expected username")
	}

	if _, err := s.DB.GetUser(context.Background(), cmd.args[0]); err != nil {
		return err
	}

	if err := s.Cfg.SetUser(cmd.args[0]); err != nil {
		return err
	}

	fmt.Println("User has been set to -", cmd.args[0])

	return nil
}

func handlerRegisters(s *state.State, cmd Command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("Expected username")
	}

	curr_time := time.Now()

	user_params := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: curr_time,
		UpdatedAt: curr_time,
		Name:      cmd.args[0],
	}

	if _, err := s.DB.CreateUser(context.Background(), user_params); err != nil {
		return err
	}

	if err := s.Cfg.SetUser(cmd.args[0]); err != nil {
		return err
	}

	fmt.Println("Successfully created and logged into user :","\nid:", user_params.ID, "\ncreated_at:", user_params.CreatedAt, "\nupdated_at:", user_params.UpdatedAt, "\nname:", user_params.Name)

	return nil
}

func handlerResets(s *state.State, cmd Command) error {
	if err := s.DB.ResetUsers(context.Background()); err != nil {
		return err
	}

	fmt.Println("Successfully cleaned all tables in the database")

	return nil
}

func handlerUsers(s *state.State, cmd Command) error {
	users, err := s.DB.GetUsers(context.Background())
	if err != nil {
		return err
	}

	if len(users) == 0 {
		fmt.Println("No users registered!")
		return nil
	}

	for _, v := range users {
		if v == s.Cfg.Curr_Username {
			fmt.Printf("* %s (current)\n", v)
			continue
		}

		fmt.Printf("* %s\n", v)
	}

	return nil
}

func scrapeFeeds(s *state.State, cmd Command) error {
	fmt.Println("Getting the next feed to be fetched...")

	feed, err := s.DB.GetNextFeedToFetch(context.Background())
	if err != nil {
		return err
	}

	c_time := time.Now()

	fmt.Printf("Fetching the feed - %s\n", feed.Name)

	feed_params := database.MarkFeedFetchedParams{
		ID:        feed.ID,
		UpdatedAt: c_time,
	}

	if err := s.DB.MarkFeedFetched(context.Background(), feed_params); err != nil {
		return err
	}

	ctx := context.Background()

	res, err := rss.FetchFeed(&ctx, feed.Url)
	if err != nil {
		return err
	}

	for _, v := range res.Channel.Item {
		s_time := time.Now()

		p_time, err := time.Parse(time.RFC1123Z, v.PubDate)
		if err != nil {
			fmt.Println("Error trying to parse time of a post -", err)
			continue
		}

		post_params := database.CreatePostParams{
			CreatedAt:   s_time,
			UpdatedAt:   s_time,
			Title:       v.Title,
			Url:         v.Link,
			Description: v.Description,
			PublishedAt: p_time,
			FeedID:      feed.ID,
		}

		if err := s.DB.CreatePost(context.Background(), post_params); err != nil {
			fmt.Println("Error trying to insert a post -", err)
		}
	}

	return nil
}

func handlerAgg(s *state.State, cmd Command) error {
	var b_time time.Duration

	if len(cmd.args) == 0 {
		return fmt.Errorf("Expected arguments")
	} else if len(cmd.args) == 1 {
		c_time, err := time.ParseDuration(cmd.args[0]) 
		if err != nil {
			return err
		}

		b_time = c_time
	} else {
		return fmt.Errorf("Too many arguments, expected 1: time")
	}

	fmt.Printf("Collecting feeds every %v\n", b_time)

	ticker := time.NewTicker(b_time)
	for range ticker.C {
		scrapeFeeds(s, cmd)
	}

	return nil
}

func clean_input(s string) string {
	if s[0] == '\'' || s[0] == '"' {
		return s[1:len(s)-1]
	}

	return s
}

func handlerAddFeed(s *state.State, cmd Command, user database.User) error {
	if len(cmd.args) < 2 {
		return fmt.Errorf("Expected name and URL of the feed")
	}
	
	url, name := clean_input(cmd.args[1]), clean_input(cmd.args[0])

	c_time := time.Now()

	feed := database.CreateFeedParams{
		CreatedAt: c_time,
		UpdatedAt: c_time, 
		Name:      name,
		Url:       url,
		UserID:    user.ID,
	}

	new_feed, err := s.DB.CreateFeed(context.Background(), feed)
	if err != nil {
		return err
	}

	follow_struct := database.CreateFeedFollowParams{
		CreatedAt: c_time,
		UpdatedAt: c_time,
		UserID:    user.ID,
		FeedID:    new_feed.ID,
	}

	if _, err := s.DB.CreateFeedFollow(context.Background(), follow_struct); err != nil {
		return err
	}

	fmt.Println("Successfully created and followed feed -", name, "; from -", url)

	return nil
}

func handlerFeeds(s *state.State, cmd Command) error {
	feeds, err := s.DB.GetFeeds(context.Background())
	if err != nil {
		return err
	}

	for _, v := range feeds {
		user, err := s.DB.GetUserByID(context.Background(), v.UserID)
		if err != nil {
			return err
		}

		fmt.Printf("#%v : Name - %s ; URL - %s ; User - %s (UID:%v)\n", v.ID, v.Name, v.Url, user.Name, v.UserID)
	}

	return nil
}

func handlerFollow(s *state.State, cmd Command, user database.User) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("Expected URL")
	}

	c_time := time.Now()

	feed, err := s.DB.GetFeedByURL(context.Background(), clean_input(cmd.args[0]))
	if err != nil {
		return err 
	}

	follow_struct := database.CreateFeedFollowParams{
		CreatedAt: c_time,
		UpdatedAt: c_time, 
		UserID:    user.ID,
		FeedID:    feed.ID,
	}

	followed, err := s.DB.CreateFeedFollow(context.Background(), follow_struct)
	if err != nil {
		return err
	}

	fmt.Printf("Successfully followed feed - %s (URL:%s) ; as user - %s\n", followed.FeedName, cmd.args[0], followed.UserName)

	return nil
}

func handlerFollowing(s *state.State, cmd Command, user database.User) error {
	feeds, err := s.DB.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return err
	}

	fmt.Printf("Currently followed feeds on User - %s (ID:%v)\n", user.Name, user.ID)
	for _, v := range feeds {
		fmt.Printf("Feed - %s ; ID - %d\n", v.FeedName, v.FeedID)
	}

	return nil
}

func handlerUnfollow(s *state.State, cmd Command, user database.User) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("Expected URL")
	}

	feed, err := s.DB.GetFeedByURL(context.Background(), clean_input(cmd.args[0]))
	if err != nil {
		return err 
	}

	unf_struct := database.UnfollowFeedParams{
		UserID: user.ID,
		FeedID: feed.ID,
	}

	if err := s.DB.UnfollowFeed(context.Background(), unf_struct); err != nil {
		return err
	}

	fmt.Printf("Successfully unfollowed feed - %s (URL:%s) ; as user - %s\n", feed.Name, feed.Url, user.Name)

	return nil
}

func handlerBrowse(s *state.State, cmd Command, user database.User) error {
	var limit int32

	if len(cmd.args) == 0 {
		limit = 2
	} else {
		limit = 0

		c_i := clean_input(cmd.args[0])
		for _, v := range c_i {
			limit = limit*10 + int32(v - '0')
		}
	}

	fmt.Printf("Displaying last %d posts from subscribed feeds...\n", limit)

	browse_params := database.GetPostsByUserParams{
		Name: user.Name,
		Limit: limit,
	}

	posts, err := s.DB.GetPostsByUser(context.Background(), browse_params)
	if err != nil {
		return err
	}

	for _, v := range posts {
		fmt.Println(v)
	}

	return nil
}

func middlewareLoggedIn(handler func(s *state.State, cmd Command, user database.User) error) func(*state.State, Command) error {
	return func(s *state.State, c Command) error {
		user, err := s.DB.GetUser(context.Background(), s.Cfg.Curr_Username)
		if err != nil {
			return err
		}

		return handler(s, c, user)
	}
}

func (c *Commands) Register_all_cmds() {
	c.register("login", handlerLogins)
	c.register("register", handlerRegisters)
	c.register("reset", handlerResets)
	c.register("users", handlerUsers)
	c.register("agg", handlerAgg)
	c.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	c.register("feeds", handlerFeeds)
	c.register("follow", middlewareLoggedIn(handlerFollow))
	c.register("following", middlewareLoggedIn(handlerFollowing))
	c.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	c.register("browse", middlewareLoggedIn(handlerBrowse))
}

func Handle_Input(new_cmds *Commands) (func(*state.State, Command) error, Command) {
	os_args := os.Args
	if len(os_args) < 2 {
		log.Fatal(fmt.Errorf("Expected arguments"))
	} 

	var cmnd Command
	
	if len(os_args) > 2 {
		cmnd = Command{
			name: os_args[1],
			args: os_args[2:],
		}
	} else {
		cmnd = Command{
			name: os_args[1],
		}
	}

	fnc, ok := new_cmds.cmd[cmnd.name];
	if !ok {
		log.Fatal(fmt.Errorf("Command doesn't exist"))
	}

	return fnc, cmnd
}
