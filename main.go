package main

import (
	"os"
	"fmt"
	"log"
	"internal/config"
)

type state struct {
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

	if err := s.cfg.SetUser(cmd.args[0]); err != nil {
		return err
	}

	fmt.Println("User has been set.")

	return nil
}



func main() {
	new_cfg, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}

	new_state := state{&new_cfg}

	new_cmds := commands{}
	new_cmds.register("login", handlerLogins)
	
	os_args := os.Args
	if len(os_args) < 2 {
		fmt.Println("Expected arguments")
		os.Exit(1)
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
		fmt.Println("Command doesn't exist")
		os.Exit(1)
	}

	if err := fnc(&new_state, cmnd); err != nil {
		log.Fatal(err)
	}
}
