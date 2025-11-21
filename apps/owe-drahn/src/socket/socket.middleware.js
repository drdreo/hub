/**
 * Socket Middleware
 *
 * Handles WebSocket communication through Redux actions.
 * Intercepts socket actions and sends them through the ConnectionManager.
 * Handles incoming messages and dispatches appropriate Redux actions.
 */

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
import { getConnectionManager } from "./ConnectionManager";
import { setSessionData } from "./session";
import {
    CONNECTION_HANDSHAKE,
    connectionStatus,
    eventMap,
    GET_ROOM_LIST,
    JOIN_ROOM,
    joinedRoom,
    joinRoomError,
    PLAYER_CHOOSE_NEXT,
    PLAYER_LOSE_LIFE,
    PLAYER_READY,
    PLAYER_ROLL_DICE,
    reconnect,
    RECONNECT,
    reconnected,
    roomError
} from "./socket.actions";

function handleJoinData(data) {
    if (data?.clientId) {
        setSessionData("clientId", data.clientId);
    } else {
        setSessionData("clientId", null);
    }

    if (data?.roomId) {
        setSessionData("roomId", data.roomId);
    } else {
        setSessionData("roomId", null);
    }
}

/**
 * Create socket middleware
 */
export function createSocketMiddleware() {
    const connectionManager = getConnectionManager();

    return store => {
        // Setup connection manager event listeners
        connectionManager.on("statusChange", status => {
            store.dispatch(connectionStatus(status));
        });

        connectionManager.on("open", () => {
            console.log("Socket connection opened");
            const state = store.getState();

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

        connectionManager.on("message", event => {
            try {
                const messages = JSON.parse(event.data);
                console.log("Received messages:", messages);

                // Handle multiple messages
                messages.forEach(message => {
                    console.log("message type:", message.type);
                    handleIncomingMessage(message, store);
                });
            } catch (error) {
                console.error("Error parsing WebSocket message:", error, event.data);
            }
        });

        return next => action => {
            switch (action.type) {
                case "GAME_LEAVE":
                    connectionManager.send("leave_room");
                    connectionManager.clearQueue();

                    // Clear session data when leaving room
                    setSessionData("clientId", null);
                    setSessionData("roomId", null);
                    break;

                case CONNECTION_HANDSHAKE:
                    connectionManager.send(eventMap[CONNECTION_HANDSHAKE], {
                        uid: action.data.uid
                    });
                    break;

                case PLAYER_READY:
                    connectionManager.send(eventMap[PLAYER_READY], action.data);
                    break;

                case PLAYER_ROLL_DICE:
                    connectionManager.send(eventMap[PLAYER_ROLL_DICE]);
                    break;

                case PLAYER_LOSE_LIFE:
                    connectionManager.send(eventMap[PLAYER_LOSE_LIFE]);
                    break;

                case PLAYER_CHOOSE_NEXT:
                    connectionManager.send(eventMap[PLAYER_CHOOSE_NEXT], action.data);
                    break;

                case GET_ROOM_LIST:
                    connectionManager.send(eventMap[GET_ROOM_LIST], action.data);
                    break;

                case JOIN_ROOM:
                    connectionManager.send(eventMap[JOIN_ROOM], action.data);
                    break;

                case RECONNECT:
                    connectionManager.send(eventMap[RECONNECT], action.data);
                    break;

                default:
                    break;
            }

            return next(action);
        };
    };
}

/**
 * Handle incoming WebSocket message
 */
function handleIncomingMessage(message, store) {
    const state = store.getState();

    switch (message.type) {
        case "join_room_result":
            handleJoinData(message.data);
            if (message.success) {
                store.dispatch(joinedRoom(message.data));
            } else {
                // Dispatch error so UI can show it
                store.dispatch(joinRoomError(message.error || "Failed to join room"));
            }
            break;

        case "reconnect_result":
            handleJoinData(message.data);
            if (message.success) {
                store.dispatch(reconnected(message.data));
            } else {
                // Reconnection failed - room might not exist anymore
                console.error("Reconnection failed:", message.error);
                store.dispatch(roomError(message.error || "Failed to reconnect to room"));
            }
            break;

        case "room_list_update":
        case "get_room_list_result":
            if (message.success) {
                // Format the data to match the expected overview format
                const overviewData = {
                    totalPlayers: message.data.reduce((sum, { playerCount }) => sum + playerCount, 0),
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
            const joinedPlayer = state.game?.players?.find(p => p.id === message.data.clientId);
            const playerName = joinedPlayer?.username || "Someone";
            store.dispatch(playerJoined(playerName));
            break;
        }

        case "client_left": {
            // Find the player name from game state using clientId
            const leftPlayer = state.game?.players?.find(p => p.id === message.data.clientId);
            const playerName = leftPlayer?.username || "Someone";
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
            store.dispatch(gameStarted());
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
            store.dispatch(feedMessage({ type: "LOST_LIFE", username: message.data.player.username }));
            break;

        default:
            console.warn("Unhandled message type:", message.type);
    }
}

export default createSocketMiddleware;
