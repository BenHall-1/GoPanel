package main

import (
	crypto_rand "crypto/rand"
	"encoding/binary"
	"fmt"
	"github.com/TicketsBot/GoPanel/app/http"
	"github.com/TicketsBot/GoPanel/app/http/endpoints/manage"
	"github.com/TicketsBot/GoPanel/config"
	"github.com/TicketsBot/GoPanel/database"
	"github.com/TicketsBot/GoPanel/messagequeue"
	"github.com/TicketsBot/GoPanel/rpc"
	"github.com/TicketsBot/GoPanel/rpc/cache"
	"github.com/TicketsBot/GoPanel/utils"
	"github.com/TicketsBot/archiverclient"
	"github.com/TicketsBot/common/premium"
	"github.com/apex/log"
	"math/rand"
	"time"
)

func main() {
	var b [8]byte
	_, err := crypto_rand.Read(b[:])
	if err == nil {
		rand.Seed(int64(binary.LittleEndian.Uint64(b[:])))
	} else {
		log.Error(err.Error())
		rand.Seed(time.Now().UnixNano())
	}

	config.LoadConfig()
	database.ConnectToDatabase()
	cache.Instance = cache.NewCache()

	manage.Archiver = archiverclient.NewArchiverClientWithTimeout(config.Conf.Bot.ObjectStore, time.Second*15, []byte(config.Conf.Bot.AesKey))

	utils.LoadEmoji()

	messagequeue.Client = messagequeue.NewRedisClient()
	go Listen(messagequeue.Client)

	rpc.PremiumClient = premium.NewPremiumLookupClient(
		premium.NewPatreonClient(config.Conf.Bot.PremiumLookupProxyUrl, config.Conf.Bot.PremiumLookupProxyKey),
		messagequeue.Client.Client,
		cache.Instance.PgCache,
		database.Client,
	)

	http.StartServer()
}

func Listen(client messagequeue.RedisClient) {
	ch := make(chan messagequeue.TicketMessage)
	go client.ListenForMessages(ch)

	for decoded := range ch {
		manage.SocketsLock.Lock()
		for _, socket := range manage.Sockets {
			if socket.Guild == decoded.GuildId && socket.Ticket == decoded.TicketId {
				if err := socket.Ws.WriteJSON(decoded); err != nil {
					fmt.Println(err.Error())
				}
			}
		}
		manage.SocketsLock.Unlock()
	}
}
