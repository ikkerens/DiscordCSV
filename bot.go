package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

type snowflake = string

const (
	guildID = "118109806723727363"
)

var (
	status     string
	statusLock sync.RWMutex
)

func main() {
	discord, err := discordgo.New("Bot " + os.Getenv("TOKEN"))
	if err != nil {
		panic(err)
	}

	discord.AddHandler(onMessage)

	if err := discord.Open(); err != nil {
		panic(err)
	}
	defer discord.Close()

	setStatus("Idle")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}

func onMessage(discord *discordgo.Session, event *discordgo.MessageCreate) {
	if !strings.HasPrefix(event.Content, "!") {
		return // Not a command
	}

	srcChannel, err := getChannel(discord, event.ChannelID)
	if err != nil {
		return // Uhm?
	}

	if srcChannel.Type != discordgo.ChannelTypeDM {
		return // Channel is not a DM channel
	}

	member, err := getGuildMember(discord, guildID, event.Author.ID)
	if err != nil {
		return // Sender is not member of the CoE server
	}

	hasRole := false
	for _, role := range member.Roles {
		if role == "197229452885884928" || role == "325679283118800907" { // Admin or SuperMod respectively
			hasRole = true
			break
		}
	}

	if !hasRole {
		return // Member does not have the SuperMod or Admin role on the CoE server
	}

	// At this point, we're authorized, let's see which command they're using.
	args := strings.Fields(strings.TrimPrefix(event.Content, "!"))
	if len(args) == 0 {
		return // Not a complete command
	}

	switch strings.ToLower(args[0]) {
	case "csv":
		csvCommand(discord, event.ChannelID, args[1:])
	case "status":
		statusCommand(discord, event.ChannelID)
	default:
		_, _ = discord.ChannelMessageSend(event.ChannelID, "I don't know that command!")
	}
}

func csvCommand(discord *discordgo.Session, srcChannelID snowflake, args []string) {
	if len(args) != 1 {
		_, _ = discord.ChannelMessageSend(srcChannelID, "Command usage: !csv <channelID>\n"+
			"Running this command will start the process to index an entire channel into a CSV file "+
			"and then upload it as an attachment in this channel.")
		return
	}

	channel, err := getChannel(discord, args[0])
	if err != nil {
		_, _ = discord.ChannelMessageSend(srcChannelID, "I'm sorry, but I can not find that channel!")
		return
	}

	go startCSVIndex(discord, srcChannelID, channel)
	_, _ = discord.ChannelMessageSend(srcChannelID, "Okay! I've started downloading that channel.\n"+
		"Please note that this process may take a while as I can only process up to 72000 messages per hour.\n"+
		"\n"+
		"~~Blame discord.~~")
}

func startCSVIndex(discord *discordgo.Session, srcChannelID snowflake, targetChannel *discordgo.Channel) {
	setStatus("Preparing to index channel #" + targetChannel.Name)
	defer setStatus("Idle")

	last := ""
	collection := make([][]string, 0)

	for {
		messages, err := discord.ChannelMessages(targetChannel.ID, 100, last, "", "")
		if err != nil {
			_, _ = discord.ChannelMessageSend(srcChannelID, "I've run into an error while indexing:\n"+err.Error())
			return
		}

		last = messages[len(messages)-1].ID

		for _, msg := range messages {
			collection = append(collection, []string{msg.Author.ID, msg.Author.Username, msg.Content})
		}

		if len(messages) < 100 {
			break // We're done with the channel!
		}

		setStatus("Indexing - " + strconv.Itoa(len(collection)) + " messages done")
		time.Sleep(5 * time.Second)
	}

	setStatus("Preparing CSV")

	// Reverse the collection
	for i, j := 0, len(collection)-1; i < j; i, j = i+1, j-1 {
		collection[i], collection[j] = collection[j], collection[i]
	}

	// Create the CSV
	output := new(bytes.Buffer)
	fmt.Fprintln(output, "AuthorID,AuthorName,Content")
	outputCSV := csv.NewWriter(output)
	outputCSV.WriteAll(collection)
	outputCSV.Flush()

	_, _ = discord.ChannelFileSend(srcChannelID, targetChannel.Name+".csv", output)
}

func setStatus(newStatus string) {
	statusLock.Lock()
	defer statusLock.Unlock()

	status = newStatus
}

func statusCommand(discord *discordgo.Session, srcChannelID snowflake) {
	statusLock.RLock()
	defer statusLock.RUnlock()

	_, _ = discord.ChannelMessageSend(srcChannelID, "Current status: "+status)
}
