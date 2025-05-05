package main

import (
	"context"
	"encoding/json"
	"fmt"
	"gameserver/games/dicegame"
	"gameserver/games/owe_drahn"
	"gameserver/games/tictactoe"
	"gameserver/internal/client"
	"gameserver/internal/events"
	"gameserver/internal/game"
	"gameserver/internal/interfaces"
	"gameserver/internal/protocol"
	"gameserver/internal/room"
	"gameserver/internal/router"
	"gameserver/internal/session"
	"github.com/getsentry/sentry-go"
	"github.com/rs/zerolog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return false // Reject requests with no origin
		}

		// Parse the origin URL
		originURL, err := url.Parse(origin)
		if err != nil {
			log.Error().Err(err).Str("origin", origin).Msg("Failed to parse origin URL")
			return false
		}

		// verify hostname
		return strings.HasSuffix(originURL.Host, "drdreo.com")
	},
}

func main() {
	rootCtx, rootCancel := context.WithCancel(context.Background())
	defer rootCancel() // Safety net - cancels if main exits unexpectedly
	initLogger()

	observeFlush := initObservability()
	defer observeFlush()

	stage := interfaces.Production
	if os.Getenv("STAGE") == "development" {
		stage = interfaces.Development

		// Allow all origins in development
		upgrader.CheckOrigin = func(r *http.Request) bool {
			return true
		}
	}

	log.Info().Str("env.STAGE", os.Getenv("STAGE")).Str("stage", string(stage)).Msg("checking environment")

	// Initialize the global session store with 5 minute expiry
	session.InitGlobalStore(300)
	eventBus := events.NewEventBus()

	gameRegistry := game.NewRegistry()
	clientManager := client.NewManager()
	roomManager := room.NewRoomManager(gameRegistry, eventBus)
	messageRouter := router.NewRouter(rootCtx, clientManager, roomManager, gameRegistry, eventBus)

	// Register all games
	tictactoe.RegisterTicTacToeGame(gameRegistry)
	dicegame.RegisterDiceGame(gameRegistry)
	if err := owe_drahn.RegisterGame(rootCtx, gameRegistry, owe_drahn.GameConfig{
		Stage:          stage,
		CredentialsDir: "apps/gameserver/internal/database/firestore/credentials",
	}); err != nil {
		log.Fatal().Err(err).Msg("Failed to register owe_drahn")
	}

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		wsHandler(w, r, messageRouter, clientManager)
	})

	// Add a simple endpoint to list available games
	http.HandleFunc("/games", func(w http.ResponseWriter, r *http.Request) {
		gamesHandler(w, gameRegistry)
	})

	// Add a new endpoint to list all rooms
	http.HandleFunc("/rooms", func(w http.ResponseWriter, r *http.Request) {
		roomHandler(w, roomManager)
	})

	port, _ := strconv.Atoi(os.Getenv("PORT"))
	addr := fmt.Sprintf(":%d", port)

	log.Info().Fields(map[string]interface{}{"port": port, "address": addr}).Msg("🎮 Server starting")

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
	}
}

func initObservability() func() {
	log.Info().Msg("initializing Sentry") // Get DSN from environment variable
	dsn := os.Getenv("SENTRY_DSN")
	if dsn == "" {
		log.Warn().Msg("SENTRY_DSN environment variable not set - skipping Sentry initialization")
		return func() {}
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:            dsn,
		Debug:          false,
		Environment:    "production",
		SendDefaultPII: true,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to init sentry")
	}
	// Flush buffered events before the program terminates.
	// Set the timeout to the maximum duration the program can afford to wait.
	// Return a cleanup function that will flush sentry
	return func() {
		sentry.Flush(2 * time.Second)
	}
}

func initLogger() {
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		return filepath.Base(file) + ":" + strconv.Itoa(line)
	}

	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = log.With().Caller().Logger().Output(zerolog.ConsoleWriter{Out: os.Stdout})

}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "Game server running"}`))
}

func wsHandler(w http.ResponseWriter, r *http.Request, router *router.Router, clientManager *client.Manager) {
	// Get interested game type info from query parameters
	gameType := r.URL.Query().Get("game")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("Error upgrading connection")
		return
	}

	c := client.NewWebsocketClient(conn, clientManager, gameType)

	// Set message handler
	c.OnMessage = func(message []byte) {
		router.HandleMessage(c, message)
	}

	// Start read/write pumps
	c.StartPumps()

	// Send welcome message
	welcomeMsg := protocol.NewSuccessResponse("welcome", interfaces.M{
		"message": "Connected to game server. Interested in game: " + gameType,
	})
	c.Send(welcomeMsg)
}

func gamesHandler(w http.ResponseWriter, gameRegistry *game.Registry) {
	w.Header().Set("Content-Type", "application/json")

	games := gameRegistry.ListGames()
	response := map[string][]string{"games": games}

	jsonData, err := json.Marshal(response)
	if err != nil {
		http.Error(w, `{"error": "failed to encode response"}`, http.StatusInternalServerError)
		return
	}

	w.Write(jsonData)
}

func roomHandler(w http.ResponseWriter, roomManager *room.RoomManager) {
	w.Header().Set("Content-Type", "application/json")

	rooms := roomManager.ListRooms()
	response := make([]interfaces.M, 0, len(rooms))

	for _, r := range rooms {
		roomInfo := interfaces.M{
			"id":          r.ID(),
			"type":        r.GameType(),
			"clientCount": len(r.Clients()),
		}
		response = append(response, roomInfo)
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		http.Error(w, `{"error": "failed to encode response"}`, http.StatusInternalServerError)
		return
	}

	w.Write(jsonData)
}
