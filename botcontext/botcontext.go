package botcontext

import (
	"github.com/TicketsBot/GoPanel/config"
	dbclient "github.com/TicketsBot/GoPanel/database"
	"github.com/TicketsBot/GoPanel/messagequeue"
	"github.com/TicketsBot/GoPanel/rpc/cache"
	"github.com/TicketsBot/database"
	"github.com/go-redis/redis"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/objects/guild"
	"github.com/rxdn/gdl/objects/member"
	"github.com/rxdn/gdl/rest"
	"github.com/rxdn/gdl/rest/ratelimit"
)

type BotContext struct {
	Token       string
	RateLimiter *ratelimit.Ratelimiter
}

func (ctx BotContext) Db() *database.Database {
	return dbclient.Client
}

func (ctx BotContext) Redis() *redis.Client {
	return messagequeue.Client.Client
}

func (ctx BotContext) IsBotAdmin(userId uint64) bool {
	for _, id := range config.Conf.Admins {
		if id == userId {
			return true
		}
	}

	return false
}

func (ctx BotContext) GetGuild(guildId uint64) (g guild.Guild, err error) {
	if guild, found := cache.Instance.GetGuild(guildId, false); found {
		return guild, nil
	}

	g, err = rest.GetGuild(ctx.Token, ctx.RateLimiter, guildId)
	if err == nil {
		go cache.Instance.StoreGuild(g)
	}

	return
}

func (ctx BotContext) GetChannel(channelId uint64) (ch channel.Channel, err error) {
	if channel, found := cache.Instance.GetChannel(channelId); found {
		return channel, nil
	}

	ch, err = rest.GetChannel(ctx.Token, ctx.RateLimiter, channelId)
	if err == nil {
		go cache.Instance.StoreChannel(ch)
	}

	return
}

func (ctx BotContext) GetGuildMember(guildId, userId uint64) (m member.Member, err error) {
	if guild, found := cache.Instance.GetMember(guildId, userId); found {
		return guild, nil
	}

	m, err = rest.GetGuildMember(ctx.Token, ctx.RateLimiter, guildId, userId)
	if err == nil {
		go cache.Instance.StoreMember(m, guildId)
	}

	return
}

func (ctx BotContext) GetGuildRoles(guildId uint64) (roles []guild.Role, err error) {
	if roles := cache.Instance.GetGuildRoles(guildId); len(roles) > 0 {
		return roles, nil
	}

	roles, err = rest.GetGuildRoles(ctx.Token, ctx.RateLimiter, guildId)
	if err == nil {
		go cache.Instance.StoreRoles(roles, guildId)
	}

	return
}
