package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/airbrake/gobrake/v5"
	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Bot struct {
	Session      *discordgo.Session
	Collection   *mongo.Collection
	ErrorReport  *gobrake.Notifier
	BotRole      string
	GuildID      string
	ChangelogURL string

	generalChannelID     string
	commandsChannelID    string
	rolesChannelID       string
	bulletinChannelID    string
	pcChannelID          string
	playstationChannelID string
	xboxChannelID        string

	roleSelfAssignMessageID string
	setupCommandID          string
	meCommandID             string
	onlineCommandID         string
	offlineCommandID        string
	showPlayersCommandID    string
}

const (
	colorWhite          = 16777215
	colorGrey           = 10070709
	colorDark           = 2895667
	colorBlurple        = 5793266
	colorGreen          = 5763719
	colorRed            = 15548997
	rdoAvatarURLPrefix  = "https://prod-cdnugc-rockstargames.akamaized.net/rdr2/pedshot/pcros/"
	rdoAvatarURLSuffix  = "/pedshot_0.jpg"
	rdoAvatarUnknownURL = "https://a.rsg.sc/s/RDR2/n/RedDeadRedemption234.png"
)

func main() {
	env := readEnv()
	bot := Bot{GuildID: env.guildID, BotRole: env.botRole, ChangelogURL: env.changelogURL}

	bot.Session = initializeBot(env)
	bot.ErrorReport = initializeErrorReport(env)

	// Database connection
	clientOptions := options.Client().
		ApplyURI("mongodb+srv://" + env.mongodbCreds + "@cluster0.w5ind.mongodb.net/?retryWrites=true&w=majority").
		SetServerAPIOptions(options.ServerAPI(options.ServerAPIVersion1))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mdbClient, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		bot.ErrorReport.Notify(err, nil)
		log.Fatal(err)
	}

	bot.Collection = mdbClient.Database(env.dbName).Collection(env.collName)

	// Create TTL index
	_, err = bot.Collection.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys:    bson.D{{Key: "expires", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(1),
		},
	)
	if err != nil {
		bot.ErrorReport.Notify(err, nil)
		log.Fatal(err)
	}

	bot.Session.AddHandler(bot.prepareServer)
	bot.Session.AddHandler(bot.registerCommands)
	bot.Session.AddHandler(bot.assignRole)
	bot.Session.AddHandler(bot.unassignRole)
	bot.Session.AddHandler(bot.userWelcome)

	bot.Session.Identify.Intents |= discordgo.IntentsAllWithoutPrivileged
	bot.Session.Identify.Intents |= discordgo.IntentGuildMembers

	// Open a websocket connection to Discord and begin listening.
	err = bot.Session.Open()
	if err != nil {
		bot.ErrorReport.Notify(err, nil)
		log.Fatalf("Error opening session: %v", err)
	}

	defer bot.ErrorReport.Close()
	defer bot.ErrorReport.NotifyOnPanic()
	defer bot.Session.Close()

	http.HandleFunc("/", healthCheck)
	http.ListenAndServe(":8080", nil)

	// Wait here until CTRL-C or other term signal is received.
	log.Println("Bot is now running.  Press CTRL-C to exit.")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-stop

	log.Println("Bot successfully shutdown.")
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	resp := make(map[string]string)
	resp["message"] = "OK"
	jsonRes, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}
	w.Write(jsonRes)
}
