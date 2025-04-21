import { feedMessage } from "../game/Feed/feed.actions";
import {
    gameError,
    gameInit,
    gameOver,
    gameStarted,
    gameUpdate,
    lostLife,
    playerLeft,
    playerUpdate,
    rolledDice
} from "../game/game.actions";
import { gameOverview } from "../home/home.actions.js";
import { getWebSocket } from "./websocket";

export default store => {
    const socket = getWebSocket();

    socket.onmessage = event => {
        const messages = JSON.parse(event.data);
        console.log("Received messages in general:", messages);

        // Handle multiple messages
        messages.forEach(message => {
            console.log("message type:", message.type);

            // out room events
            switch (message.type) {
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
                // ... other cases
            }
        });
    };

    socket.onopen = () => console.log("Socket connected!");
    socket.onclose = () => console.log("Socket disconnected!");
};

export const initializeGameSocketListeners = (socket, dispatch) => {
    socket.onmessage = event => {
        const messages = JSON.parse(event.data);
        console.log("Received messages in game listeners:", messages);
        // Handle multiple messages
        messages.forEach(message => {
            console.log("message type:", message.type);

            switch (message.type) {
                case "gameInit":
                    dispatch(gameInit(message.data));
                    break;
                case "gameStarted":
                    dispatch(gameStarted(message.data));
                    break;
                case "game_state":
                    dispatch(gameUpdate(message.data));
                    break;
                case "gameOver":
                    dispatch(gameOver(message.data.winner));
                    break;
                case "gameError":
                    dispatch(gameError(message.data));
                    break;
                case "playerUpdate":
                    dispatch(playerUpdate(message.data));
                    break;
                case "playerLeft":
                    dispatch(playerLeft(message.data.username));
                    break;
                case "rolledDice":
                    dispatch(rolledDice(message.data));
                    break;
                case "lostLife":
                    dispatch(lostLife());
                    dispatch(
                        feedMessage({ type: "LOST_LIFE", username: message.data.player.username })
                    );
                    break;
                case "lost":
                    // dispatch(playerLost(data.player.id));
                    // dispatch(feedMessage({type: "LOST", username: data.player.username, dice: data.dice, total: data.total}));
                    break;

                default:
                    console.warn("Unhandled message type:", message.type);
            }
        });
    };
};
