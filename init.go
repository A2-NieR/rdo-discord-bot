package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/airbrake/gobrake/v5"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

type Env struct {
	environment, botToken, botRole, guildID, changelogURL, airbrakeKey, mongodbCreds, dbName, collName string
	airbrakeID                                                                                         int64
}

func readEnv() *Env {
	fmt.Println("================================================================================")
	log.Println("Reading environment variables...")
	if _, err := os.Stat(".env"); err == nil {
		envs, err := godotenv.Read(".env")
		if err != nil {
			log.Fatal("Error loading .env file")
		}

		developmentEnvironment := Env{environment: "DEVELOPMENT", botToken: envs["BOT_TOKEN"], botRole: envs["BOT_ROLE"], guildID: envs["DEV_GUILD_ID"], changelogURL: envs["CHANGELOG"], airbrakeKey: envs["AIRBRAKE_KEY"], mongodbCreds: envs["MONGODB_CREDS"], dbName: envs["DB"], collName: "players"}

		airbrakeIDString := envs["AIRBRAKE_ID"]
		airbrakeIDToInt, _ := strconv.Atoi(airbrakeIDString)
		developmentEnvironment.airbrakeID = int64(airbrakeIDToInt)

		return &developmentEnvironment
	} else {
		productionEnvironment := Env{environment: "PRODUCTION", botToken: os.Getenv("BOT_TOKEN"), botRole: os.Getenv("BOT_ROLE"), guildID: os.Getenv("GUILD_ID"), changelogURL: os.Getenv("CHANGELOG"), airbrakeKey: os.Getenv("AIRBRAKE_KEY"), mongodbCreds: os.Getenv("MONGODB_CREDS"), dbName: os.Getenv("DB"), collName: "players"}

		airbrakeIDToInt, _ := strconv.Atoi(os.Getenv("AIRBRAKE_ID"))
		productionEnvironment.airbrakeID = int64(airbrakeIDToInt)

		return &productionEnvironment
	}
}

func initializeBot(e *Env) *discordgo.Session {
	log.Printf("Starting bot in %s mode...", e.environment)

	s, err := discordgo.New("Bot " + e.botToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}

	return s
}

func initializeErrorReport(e *Env) *gobrake.Notifier {
	log.Println("Register error logging...")
	return gobrake.NewNotifierWithOptions(&gobrake.NotifierOptions{
		ProjectId:   e.airbrakeID,
		ProjectKey:  e.airbrakeKey,
		Environment: strings.ToLower(e.environment),
	})
}
