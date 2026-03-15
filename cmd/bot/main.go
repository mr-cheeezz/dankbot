package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	configPath := flag.String("config", "configs/example.ini", "path to config file")
	flag.Parse()

	ctx, stop := signalContext()
	defer stop()

	app, err := newApplication(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create bot application: %v\n", err)
		os.Exit(1)
	}

	if err := app.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "bot exited with error: %v\n", err)
		os.Exit(1)
	}
}
