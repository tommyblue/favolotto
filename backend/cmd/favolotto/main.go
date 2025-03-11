package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/signal"

	favolotto "github.com/tommyblue/favolotto/backup"
)

var (
	configFlag = flag.String("config", "config.json", "config file path")
)

func main() {
	flag.Parse()

	if *configFlag == "" {
		log.Println("config file is required")
		flag.Usage()

		return
	}

	if _, err := os.Stat(*configFlag); os.IsNotExist(err) {
		log.Fatalf("config file %s does not exist", *configFlag)
	}

	content, err := os.ReadFile(*configFlag)
	if err != nil {
		log.Fatal(err)
	}

	var config favolotto.Config
	if err := json.Unmarshal(content, &config); err != nil {
		panic(err)
	}

	ctx, end := signal.NotifyContext(context.Background(), os.Interrupt)
	defer end()

	f := favolotto.New(config)
	if err := f.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
