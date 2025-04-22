import {
    CONNECTION_HANDSHAKE,
    eventMap,
    GET_ROOM_LIST,
    JOIN_ROOM,
    PLAYER_CHOOSE_NEXT,
    PLAYER_LOSE_LIFE,
    PLAYER_READY,
    PLAYER_ROLL_DICE,
    RECONNECT
} from "./socket.actions";

import { connectWebSocket, getWebSocket } from "./websocket";

const initialState = { socket: connectWebSocket() };

const sendMessage = (socket, type, payload) => {
    if (socket.readyState !== WebSocket.OPEN) {
        console.error("WebSocket is not connected");
        return;
    }
    const message = {
        type: type,
        data: payload
    };
    socket.send(JSON.stringify(message));
};

const socketReducer = (state = initialState, action) => {
    const socket = getWebSocket();
    if (!socket) {
        console.error("WebSocket not initialized!");
        return state;
    }

    switch (action.type) {
        case "GAME_RESET":
            sendMessage(socket, "leave_room");
            return state;
        case CONNECTION_HANDSHAKE:
            sendMessage(socket, eventMap[action.type], {
                playerId: sessionStorage.getItem("playerId"),
                room: action.data.room,
                uid: action.data.uid
            });
            return state;
        case PLAYER_READY:
            sendMessage(socket, eventMap[action.type], action.data);
            return state;
        case PLAYER_ROLL_DICE:
            sendMessage(socket, eventMap[PLAYER_ROLL_DICE]);
            return state;
        case PLAYER_LOSE_LIFE:
            sendMessage(socket, eventMap[PLAYER_LOSE_LIFE]);
            return state;
        case PLAYER_CHOOSE_NEXT:
            sendMessage(socket, eventMap[PLAYER_CHOOSE_NEXT], action.data);
            return state;
        case GET_ROOM_LIST:
            sendMessage(socket, eventMap[GET_ROOM_LIST], action.data);
            return state;
        case JOIN_ROOM:
            sendMessage(socket, eventMap[JOIN_ROOM], action.data);
            return state;
        case RECONNECT:
            sendMessage(socket, eventMap[JOIN_ROOM], action.data);
            return state;
        default:
            return state;
    }
};

export default socketReducer;
