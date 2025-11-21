import { feedMessage } from "../game/Feed/feed.actions";
import {
    gameInit,
    gameOver,
    gameStarted,
    gameUpdate,
    lostLife,
    playerJoined,
    playerLeft,
    playerUpdate,
    rolledDice
} from "../game/game.actions";
import { gameOverview } from "../home/home.actions.js";
import { connectionStatus, joinedRoom, reconnect, reconnected } from "./socket.actions";
import { getWebSocket, updateLastMessageTime } from "./websocket";

function handleJoinData(data) {
    if (!data?.clientId) {
        sessionStorage.removeItem("clientId");
        localStorage.removeItem("clientId_backup");
    } else {
        sessionStorage.setItem("clientId", data.clientId);
        // Store in localStorage as fallback for mobile browsers
        localStorage.setItem("clientId_backup", data.clientId);
        localStorage.setItem("clientId_timestamp", Date.now().toString());
    }

    if (!data?.roomId) {
        sessionStorage.removeItem("roomId");
        localStorage.removeItem("roomId_backup");
    } else {
        sessionStorage.setItem("roomId", data.roomId);
        // Store in localStorage as fallback for mobile browsers
        localStorage.setItem("roomId_backup", data.roomId);
    }
}

export default store => {
    const socket = getWebSocket();

    socket.addEventListener("open", () => {
        console.warn("Socket connection opened");
        const state = store.getState();
        store.dispatch(connectionStatus(WebSocket.OPEN));

        // Only attempt reconnection if we have stored session data
        if (state.socket.clientId && state.socket.roomId) {
            console.log(
                "Attempting to reconnect with stored session:",
                state.socket.clientId,
                state.socket.roomId
            );
            store.dispatch(reconnect(state.socket.clientId, state.socket.roomId));
        } else {
            console.log("No stored session, waiting for join/handshake");
        }
    });

    socket.addEventListener("close", () => {
        console.warn("Socket connection closed");
        store.dispatch(connectionStatus(WebSocket.CLOSED));
    });

    socket.onmessage = event => {
        // Update health check timestamp
        updateLastMessageTime();

        const messages = JSON.parse(event.data);
        console.log("Received messages in general:", messages);

        // Handle multiple messages
        messages.forEach(message => {
            console.log("message type:", message.type);

            // out room events
            switch (message.type) {
                case "join_room_result":
                    handleJoinData(message.data);
                    if (message.success) {
                        store.dispatch(joinedRoom(message.data));
                    }
                    break;

                case "reconnect_result":
                    handleJoinData(message.data);
                    if (message.success) {
                        store.dispatch(reconnected(message.data));
                    }
                    break;

                case "room_list_update":
                case "get_room_list_result":
                    if (message.success) {
                        // Format the data to match the expected overview format
                        const overviewData = {
                            totalPlayers: message.data.reduce(
                                (sum, { playerCount }) => sum + playerCount,
                                0
                            ),
                            rooms: message.data.map(({ roomId, started }) => ({
                                room: roomId,
                                started
                            }))
                        };
                        store.dispatch(gameOverview(overviewData));
                    } else {
                        console.error("Error fetching room list:", message.error);
                    }
                    break;

                case "client_joined": {
                    // Find the player name from game state using clientId
                    const state = store.getState();
                    const joinedPlayer = state.game?.players?.find(p => p.id === message.data.clientId);
                    const playerName = joinedPlayer?.username || "Someone";
                    store.dispatch(playerJoined(playerName));
                    break;
                }
                case "client_left": {
                    // Find the player name from game state using clientId
                    const state = store.getState();
                    const joinedPlayer = state.game?.players?.find(p => p.id === message.data.clientId);

                    const playerName = joinedPlayer?.username || "Someone";
                    store.dispatch(playerLeft(playerName));
                    break;
                }

                case "playerLeft":
                    store.dispatch(playerLeft(message.data.playerName ?? "Someone"));
                    break;

                case "gameInit":
                    store.dispatch(gameInit(message.data));
                    break;
                case "gameStarted":
                    store.dispatch(gameStarted(message.data));
                    break;
                case "game_state":
                    store.dispatch(gameUpdate(message.data));
                    break;
                case "gameOver":
                    store.dispatch(gameOver(message.data.winner));
                    break;
                case "playerUpdate":
                    store.dispatch(playerUpdate(message.data));
                    break;
                case "rolledDice":
                    store.dispatch(rolledDice(message.data));
                    break;
                case "lostLife":
                    store.dispatch(lostLife());
                    store.dispatch(
                        feedMessage({ type: "LOST_LIFE", username: message.data.player.username })
                    );
                    break;
                default:
                    console.warn("Unhandled message type:", message.type);
            }
        });
    };
};
