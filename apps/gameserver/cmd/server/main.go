package main

import (
	"encoding/json"
	"github.com/drdreo/hub/gameserver/games/sample"
	"github.com/drdreo/hub/gameserver/internal/client"
	"github.com/drdreo/hub/gameserver/internal/game"
	"github.com/drdreo/hub/gameserver/internal/room"
	"github.com/drdreo/hub/gameserver/internal/router"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for dev
	},
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	var port = os.Getenv("PORT")

	// Initialize game registry
	gameRegistry := game.NewRegistry()

	// Initialize room manager
	roomManager := room.NewRoomManager(gameRegistry)

	// Initialize router
	messageRouter := router.NewRouter(roomManager, gameRegistry)

	// Register example game (to be replaced with real game implementations)
	sample.RegisterTicTacToeGame(gameRegistry)

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

	addr := "0.0.0.0:" + port
	log.Printf("🎮 server starting on %s ...\n", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
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
		log.Println("Error upgrading connection:", err)
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
	c.Send([]byte(`{"type":"welcome","message":"Connected to game server"}`))

	log.Printf("Client connected: %s", c.ID())
}
