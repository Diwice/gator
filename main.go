package main

import (
	"os"
	"fmt"
	"log"
	"time"
	"context"
	"database/sql"
	"gator/internal/rss"
	"gator/internal/config"
	"gator/internal/database"
	"github.com/google/uuid"

	_ "github.com/lib/pq"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	cmd map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	get_cmd := c.cmd[cmd.name]

	if err := get_cmd(s, cmd); err != nil {
		return err
	}

	return nil
}

func (c *commands) register(name string, f func(*state, command) error) {
	if c.cmd == nil {
		c.cmd = make(map[string]func(*state, command) error)
	}

	c.cmd[name] = f
}

func handlerLogins(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("Expected username")
	}

	if _, err := s.db.GetUser(context.Background(), cmd.args[0]); err != nil {
		return err
	}

	if err := s.cfg.SetUser(cmd.args[0]); err != nil {
		return err
	}

	fmt.Println("User has been set to -", cmd.args[0])

	return nil
}

func handlerRegisters(s *state, cmd command) error {
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

	if _, err := s.db.CreateUser(context.Background(), user_params); err != nil {
		return err
	}

	if err := s.cfg.SetUser(cmd.args[0]); err != nil {
		return err
	}

	fmt.Println("Successfully created and logged into user :","\nid:", user_params.ID, "\ncreated_at:", user_params.CreatedAt, "\nupdated_at:", user_params.UpdatedAt, "\nname:", user_params.Name)

	return nil
}

func handlerResets(s *state, cmd command) error {
	if err := s.db.ResetUsers(context.Background()); err != nil {
		return err
	}

	fmt.Println("Successfully resetted the users table")

	return nil
}

func handlerUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}

	if len(users) == 0 {
		fmt.Println("No users registered!")
		return nil
	}

	for _, v := range users {
		if v == s.cfg.Curr_Username {
			fmt.Printf("* %s (current)\n", v)
			continue
		}

		fmt.Printf("* %s\n", v)
	}

	return nil
}

func handlerAgg(s *state, cmd command) error {
	var url string

	if len(cmd.args) == 0 {
		return fmt.Errorf("Expected arguments")
	} else if len(cmd.args) == 1 {
		url = cmd.args[0]
	} else {
		url = cmd.args[1]
	}

	ctx := context.Background()

	res, err := rss.FetchFeed(&ctx, url)
	if err != nil {
		return err
	}

	fmt.Println(res)

	return nil
}

func clean_input(s string) string {
	if s[0] == '\'' || s[0] == '"' {
		return s[1:len(s)-1]
	}

	return s
}

func handlerAddFeed(s *state, cmd command) error {
	if len(cmd.args) < 2 {
		return fmt.Errorf("Expected name and url of the feed")
	}
	
	url, name := clean_input(cmd.args[1]), clean_input(cmd.args[0])

	user, err := s.db.GetUser(context.Background(), s.cfg.Curr_Username)
	if err != nil {
		return err
	}

	c_time := time.Now()

	feed := database.CreateFeedParams{
		CreatedAt: c_time,
		UpdatedAt: c_time, 
		Name:      name,
		Url:       url,
		UserID:    user.ID,
	}

	if _, err := s.db.CreateFeed(context.Background(), feed); err != nil {
		return err
	}

	fmt.Println("Successfully created feed -", name, "; from -", url)

	return nil
}

func (c *commands) register_all_cmds() {
	c.register("login", handlerLogins)
	c.register("register", handlerRegisters)
	c.register("reset", handlerResets)
	c.register("users", handlerUsers)
	c.register("agg", handlerAgg)
	c.register("addfeed", handlerAddFeed)
}

func handle_input(new_cmds *commands) (func(*state, command) error, command) {
	os_args := os.Args
	if len(os_args) < 2 {
		log.Fatal(fmt.Errorf("Expected arguments"))
	} 

	var cmnd command
	
	if len(os_args) > 2 {
		cmnd = command{
			name: os_args[1],
			args: os_args[2:],
		}
	} else {
		cmnd = command{
			name: os_args[1],
		}
	}

	fnc, ok := new_cmds.cmd[cmnd.name];
	if !ok {
		log.Fatal(fmt.Errorf("Command doesn't exist"))
	}

	return fnc, cmnd
}

func main() {
	new_cfg, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("postgres", new_cfg.DB_URL)
	if err != nil {
		log.Fatal(err)
	}

	dbQueries := database.New(db)

	new_state := state{dbQueries, &new_cfg}

	new_cmds := commands{}
	new_cmds.register_all_cmds()
	
	fnc, cmnd := handle_input(&new_cmds)

	if err := fnc(&new_state, cmnd); err != nil {
		log.Fatal(err)
	}
}
