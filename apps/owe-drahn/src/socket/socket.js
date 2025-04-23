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
import { joinedRoom, reconnect, reconnected } from "./socket.actions";
import { getWebSocket } from "./websocket";

function handleJoinData(data) {
    if (!data?.clientId) {
        sessionStorage.removeItem("clientId");
    } else {
        sessionStorage.setItem("clientId", data.clientId);
    }

    if (!data?.roomId) {
        sessionStorage.removeItem("roomId");
    } else {
        sessionStorage.setItem("roomId", data.roomId);
    }
}

export default store => {
    const socket = getWebSocket();

    socket.addEventListener("open", () => {
        console.warn("Socket connection opened");
        const state = store.getState();
        store.dispatch(reconnect(state.socket.clientId, state.socket.roomId));
    });

    socket.onmessage = event => {
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
                case "gameError":
                    store.dispatch(gameError(message.data));
                    break;
                case "playerUpdate":
                    store.dispatch(playerUpdate(message.data));
                    break;
                case "playerLeft":
                    store.dispatch(playerLeft(message.data.username));
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
//
// export const initializeGameSocketListeners = (socket, dispatch) => {
//     socket.onmessage = event => {
//         const messages = JSON.parse(event.data);
//         console.log("Received messages in game listeners:", messages);
//         // Handle multiple messages
//         messages.forEach(message => {
//             console.log("message type:", message.type);
//
//             switch (message.type) {
//                 case "gameInit":
//                     dispatch(gameInit(message.data));
//                     break;
//                 case "gameStarted":
//                     dispatch(gameStarted(message.data));
//                     break;
//                 case "game_state":
//                     dispatch(gameUpdate(message.data));
//                     break;
//                 case "gameOver":
//                     dispatch(gameOver(message.data.winner));
//                     break;
//                 case "gameError":
//                     dispatch(gameError(message.data));
//                     break;
//                 case "playerUpdate":
//                     dispatch(playerUpdate(message.data));
//                     break;
//                 case "playerLeft":
//                     dispatch(playerLeft(message.data.username));
//                     break;
//                 case "rolledDice":
//                     dispatch(rolledDice(message.data));
//                     break;
//                 case "lostLife":
//                     dispatch(lostLife());
//                     dispatch(
//                         feedMessage({ type: "LOST_LIFE", username: message.data.player.username })
//                     );
//                     break;
//                 case "lost":
//                     // dispatch(playerLost(data.player.id));
//                     // dispatch(feedMessage({type: "LOST", username: data.player.username, dice: data.dice, total: data.total}));
//                     break;
//
//                 default:
//                     console.warn("Unhandled message type:", message.type);
//             }
//         });
//     };
// };
