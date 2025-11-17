package main

import (
	"log"
	"database/sql"
	"gator/internal/state"
	"gator/internal/config"
	"gator/internal/handlers"
	"gator/internal/database"

	_ "github.com/lib/pq"
)

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

	new_state := state.State{dbQueries, &new_cfg}

	new_cmds := handlers.Commands{}
	new_cmds.Register_all_cmds()
	
	fnc, cmnd := handlers.Handle_Input(&new_cmds)

	if err := fnc(&new_state, cmnd); err != nil {
		log.Fatal(err)
	}
}
