package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mr-cheeezz/dankbot/pkg/common/config"
)

func main() {
	configPath := flag.String("config", "configs/example.ini", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config %q: %v\n", *configPath, err)
		os.Exit(1)
	}

	fmt.Printf("config loaded successfully for bot %s\n", cfg.Main.BotID)
}
