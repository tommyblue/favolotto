package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/tommyblue/favolotto"
)

var (
	version     = "--- set at buildtime ---"
	help        = flag.Bool("help", false, "show command help")
	configFlag  = flag.String("config", "config.json", "config file path")
	showVersion = flag.Bool("version", false, "show command version")
)

func main() {
	flag.Parse()

	if *help {
		flag.Usage()
		return
	}
	if *showVersion {
		fmt.Println(version)
		return
	}

	if *configFlag == "" {
		log.Println("config file is required")
		flag.Usage()

		return
	}

	if _, err := os.Stat(*configFlag); os.IsNotExist(err) {
		flag.Usage()
		log.Fatalf("config file %s does not exist", *configFlag)
	}

	content, err := os.ReadFile(*configFlag)
	if err != nil {
		log.Fatal("error reading config file: ", err)
	}

	var config favolotto.Config
	if err := json.Unmarshal(content, &config); err != nil {
		log.Fatal("error unmarshalling config file: ", err)
	}

	ctx, end := signal.NotifyContext(context.Background(), os.Interrupt)
	defer end()

	f := favolotto.New(config)
	if err := f.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
