package main

import (
	"fmt"
	"internal/config"
)

func main() {
	new_cfg, err := config.Read()
	if err != nil {
		fmt.Println(err)
	}
	
	new_cfg.SetUser("denis")

	re_cfg, err := config.Read()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(re_cfg)
}
