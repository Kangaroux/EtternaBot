package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Kangaroux/etternabot/etterna"
)

var apiKey string

func init() {
	flag.StringVar(&apiKey, "key", "", "api key for the EtternaOnline api")
	flag.Parse()
}

func main() {
	if apiKey == "" {
		flag.Usage()
		os.Exit(1)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		api := etterna.New(apiKey)

		for {
			fmt.Println("Getting...")
			u, err := api.GetUsername("jesse")

			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			} else {
				fmt.Printf("%+v\n", u)
			}

			<-time.After(30 * time.Second)
		}
	}()

	<-quit
	fmt.Println("Stopping")
}
