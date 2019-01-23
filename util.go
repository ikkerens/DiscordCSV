package main

/*
All functions in this file are copied from https://github.com/ikkerens/gophbot.
Implementation based on https://discordapp.com/developers/docs/topics/permissions
Ported by Rens "Ikkerens" Rikkerink
*/

import (
	"github.com/bwmarrin/discordgo"
)

// getChannel attempts to get a channel instance from the shard state cache, and if none exists, attempts to obtain it
// from the Discord API. Will err if the channel does not exist or the bot is not in the guild it belongs to.
func getChannel(discord *discordgo.Session, channelID snowflake) (*discordgo.Channel, error) {
	channel, err := discord.State.Channel(channelID)
	if err == nil {
		return channel, nil
	}

	channel, err = discord.Channel(channelID)
	if err != nil {
		return nil, err
	}
	discord.State.ChannelAdd(channel)

	return channel, err
}

// getGuildMember will attempt to obtain a member instance for this guild member.
// Will err if this user is not a member of this guild
func getGuildMember(discord *discordgo.Session, guildID, userID snowflake) (*discordgo.Member, error) {
	member, err := discord.State.Member(guildID, userID)
	if err == nil {
		return member, nil
	}

	member, err = discord.GuildMember(guildID, userID)
	if err != nil {
		return nil, err
	}
	discord.State.MemberAdd(member)

	return member, nil
}

// getUserName is a convenience function that will attempt to obtain the users username and discriminator for logging
// otherwise it will just return the user ID
func getUserName(discord *discordgo.Session, guildID, userID snowflake) string {
	member, err := discord.State.Member(guildID, userID)
	if err != nil {
		return "ID " + userID
	}

	return member.User.String()
}
