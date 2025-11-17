package state

import (
	"gator/internal/config"
	"gator/internal/database"
)

// Top 10 dependancies. Numero 10 :

type State struct {
	DB  *database.Queries
	Cfg *config.Config
}
