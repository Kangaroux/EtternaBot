package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/Kangaroux/etternabot/bot"
	"github.com/bwmarrin/discordgo"
	"github.com/jmoiron/sqlx"
)

var (
	botToken      string
	etternaAPIKey string
)

func init() {
	flag.StringVar(&botToken, "token", "", "discord bot token")
	flag.StringVar(&etternaAPIKey, "etterna-key", "", "api key for the EtternaOnline api")
	flag.Parse()
}

func main() {
	if botToken == "" || etternaAPIKey == "" {
		flag.Usage()
		os.Exit(1)
	}

	dg, err := discordgo.New("Bot " + botToken)

	if err != nil {
		fmt.Println("Failed to create discord session:", err)
		os.Exit(1)
	}

	db, err := connectDB(getenv("DATABASE_HOST"),
		getenv("POSTGRES_DB"),
		getenv("POSTGRES_USER"),
		getenv("POSTGRES_PASSWORD"))

	if err != nil {
		fmt.Println("Failed to connect to database:", err)
		os.Exit(1)
	}

	bot.New(dg, db, etternaAPIKey)

	if err := dg.Open(); err != nil {
		fmt.Println("Failed to open discord connection:", err)
		os.Exit(1)
	}

	fmt.Println("Bot is now running. Press CTRL-C to exit.")

	wait()
	dg.Close()
}

func connectDB(host, dbName, user, password string) (db *sqlx.DB, err error) {
	connStr := fmt.Sprintf("host=%s dbname=%s user=%s password=%s sslmode=disable", host, dbName, user, password)

	for retries := 0; retries < 3; retries++ {
		db, err = sqlx.Connect("postgres", connStr)

		if err == nil {
			return db, nil
		}

		// Retry in 1 second
		<-time.After(1 * time.Second)
	}

	if err == nil {
		err = errors.New("failed after 3 retries")
	}

	return nil, err
}

func getenv(key string) string {
	val, exists := os.LookupEnv(key)

	if !exists {
		fmt.Println("Environment var is not defined:", key)
		os.Exit(1)
	}

	return val
}

func wait() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}
