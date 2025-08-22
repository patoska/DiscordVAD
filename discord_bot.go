package main

import (
	"fmt"
	"os"
	"time"
	"net/http"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

var (
	Token          = ""
	GuildID        = ""
	ChannelID      = ""
	upgrader       = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	users          = make(map[string]*User)
	ssrcMap        = make(map[uint32]string)
	usersMux       sync.Mutex
	lastUpdate     = time.Now()

	session        *discordgo.Session
	voiceConn      *discordgo.VoiceConnection
	reconnectMutex sync.Mutex
)

type User struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Avatar       string    `json:"avatar"`
	Speaking     bool      `json:"speaking"`
	LastActivity time.Time `json:"last_activity"`
}

func handleVoice(c chan *discordgo.Packet) {
	for p := range c {
		now := time.Now()
		diff := now.Sub(lastUpdate)
		if diff > 200 * time.Millisecond {
			lastUpdate = now

			if user_id, ok := ssrcMap[uint32(p.SSRC)]; ok {
				users[user_id].LastActivity = now
				broadcast()
			}
		}
	}
}

func setupHandlers(discordGoSession *discordgo.Session) {
	discordGoSession.AddHandler(func(s *discordgo.Session, vs *discordgo.VoiceStateUpdate) {
		if vs.ChannelID == ChannelID {
			user, _ := s.User(vs.UserID)
			if user != nil {
				usersMux.Lock()
				users[user.ID] = &User{
					ID:     user.ID,
					Name:   user.Username,
					Avatar: user.AvatarURL("32"),
				}
				usersMux.Unlock()
			}
		} else {
			usersMux.Lock()
			delete(users, vs.UserID)
			usersMux.Unlock()
		}
		broadcast()
	})
}


func connectDiscord() (*discordgo.Session, error) {
	session, err := discordgo.New("Bot " + Token)
	if err != nil {
		return nil, err
		fmt.Println("")
	}

	session.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildVoiceStates)

	setupHandlers(session)

	if err := session.Open(); err != nil {
		return nil, err
	}

	return session, nil
}

func joinVoice(discordGoSession *discordgo.Session) (*discordgo.VoiceConnection, error) {
	vc, err := session.ChannelVoiceJoin(GuildID, ChannelID, true, false)
	
	if err != nil {
		return nil, err
	}

	vc.AddHandler(func(vc *discordgo.VoiceConnection, vs *discordgo.VoiceSpeakingUpdate) {
		usersMux.Lock()
		if _, ok := ssrcMap[uint32(vs.SSRC)]; !ok {
			ssrcMap[uint32(vs.SSRC)] = vs.UserID
		}
		usersMux.Unlock()
		broadcast()
	})

	return vc, nil
}

func ensureConnected() {
	reconnectMutex.Lock()
	defer reconnectMutex.Unlock()

	if session == nil {
		var err error
		for {
			fmt.Println("Connecting to Discord...")
			session, err = connectDiscord()
			if err == nil {
				fmt.Println("Connected to Discord API ✅")
				break
			}
			fmt.Println("Discord connect failed, retrying in 5s:", err)
			time.Sleep(5 * time.Second)
		}
	}

	if voiceConn == nil || !voiceConn.Ready {
		var err error
		for {
			fmt.Println("Joining voice channel...")
			voiceConn, err = joinVoice(session)
			if err == nil {
				fmt.Println("Joined voice channel ✅")
				break
			}
			fmt.Println("Voice join failed, retrying in 5s:", err)
			time.Sleep(5 * time.Second)
		}
	}
}

func connectBot()(bool){
	err := godotenv.Load()
  if err != nil {
		fmt.Println("Error loading .env file")
		return false
  }

	Token = os.Getenv("DISCORD_TOKEN")
	GuildID = os.Getenv("DISCORD_GUILD_ID")
	ChannelID = os.Getenv("DISCORD_VOICE_CHANNEL_ID")

	if Token == "" || GuildID == "" || ChannelID == "" {
		fmt.Println("Missing bot information, exiting...")
		return false
	}

	ensureConnected()

	go func() {
		for {
			time.Sleep(10 * time.Second)
			if session == nil || session.DataReady == false {
				fmt.Println("Session lost, reconnecting...")
				session = nil
				voiceConn = nil
				ensureConnected()
			}
			if voiceConn == nil || !voiceConn.Ready {
				fmt.Println("Voice connection lost, reconnecting...")
				voiceConn = nil
				ensureConnected()
			}
		}
	}()

	go func() {
			fmt.Println("voiceConn.Receive...")
			handleVoice(voiceConn.OpusRecv)
	}()

	return true
}

func disconnectBot(){
	voiceConn.Disconnect()
	voiceConn.Close()
	session.Close()
}
