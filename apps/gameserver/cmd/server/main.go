package main

import (
	"encoding/json"
	"fmt"
	"gameserver/games/dicegame"
	"gameserver/games/tictactoe"
	"gameserver/internal/client"
	"gameserver/internal/game"
	"gameserver/internal/interfaces"
	"gameserver/internal/protocol"
	"gameserver/internal/room"
	"gameserver/internal/router"
	"gameserver/internal/session"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for dev
	},
}

func main() {
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		return filepath.Base(file) + ":" + strconv.Itoa(line)
	}

	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = log.With().Caller().Logger().Output(zerolog.ConsoleWriter{Out: os.Stdout})

	// Initialize the global session store with 5 minute expiry
	session.InitGlobalStore(300)

	// Initialize game registry
	gameRegistry := game.NewRegistry()

	// Initialize room manager
	roomManager := room.NewRoomManager(gameRegistry)

	// Initialize router
	messageRouter := router.NewRouter(roomManager, gameRegistry)

	// Register example game (to be replaced with real game implementations)
	tictactoe.RegisterTicTacToeGame(gameRegistry)
	dicegame.RegisterDiceGame(gameRegistry)

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		wsHandler(w, r, messageRouter)
	})

	// Add a simple endpoint to list available games
	http.HandleFunc("/games", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		games := gameRegistry.ListGames()
		response := map[string][]string{"games": games}

		jsonData, err := json.Marshal(response)
		if err != nil {
			http.Error(w, `{"error": "failed to encode response"}`, http.StatusInternalServerError)
			return
		}

		w.Write(jsonData)
	})

	port, _ := strconv.Atoi(os.Getenv("PORT"))
	addr := fmt.Sprintf(":%d", port)

	log.Info().Fields(map[string]interface{}{"port": port, "address": addr}).Msg("🎮 Server starting")

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
	}
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

func wsHandler(w http.ResponseWriter, r *http.Request, router *router.Router) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("Error upgrading connection")
		return
	}

	// Create new client
	c := client.NewClient(conn)

	// Set message handler
	c.OnMessage = func(message []byte) {
		router.HandleMessage(c, message)
	}

	// Start read/write pumps
	c.StartPumps()

	// Send welcome message
	welcomeMsg := protocol.NewSuccessResponse("welcome", interfaces.M{
		"message": "Connected to game server",
	})
	c.Send(welcomeMsg)
}
