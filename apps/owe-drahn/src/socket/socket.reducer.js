import {
    CONNECTION_HANDSHAKE,
    eventMap,
    GET_ROOM_LIST,
    JOIN_ROOM,
    JOINED_ROOM,
    PLAYER_CHOOSE_NEXT,
    PLAYER_LOSE_LIFE,
    PLAYER_READY,
    PLAYER_ROLL_DICE,
    RECONNECT,
    RECONNECTED,
    RESET_RECONNECTED
} from "./socket.actions";

import { connectWebSocket, getWebSocket } from "./websocket";

const initialState = {
    socket: connectWebSocket(),
    clientId: sessionStorage.getItem("clientId"),
    roomId: sessionStorage.getItem("roomId"),
    connectionStatus: WebSocket.CLOSED,
};

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
        case "CONNECTION_STATUS":
            return {
                ...state,
                connectionStatus: action.data.status
            };
        case "GAME_LEAVE":
            sendMessage(socket, "leave_room");
            return state;
        case CONNECTION_HANDSHAKE:
            sendMessage(socket, eventMap[CONNECTION_HANDSHAKE], {
                clientId: sessionStorage.getItem("clientId"),
                room: action.data.room,
                uid: action.data.uid
            });
            return state;
        case PLAYER_READY:
            sendMessage(socket, eventMap[PLAYER_READY], action.data);
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
        case JOINED_ROOM:
            return {
                ...state,
                joinedRoom: true,
                clientId: action.data.clientId,
                roomId: action.data.roomId
            };
        case RECONNECT:
            sendMessage(socket, eventMap[RECONNECT], action.data);
            return state;
        case RESET_RECONNECTED:
            return {
                ...state,
                reconnected: false
            };
        case RECONNECTED:
            return {
                ...state,
                reconnected: true,
                clientId: action.data.clientId,
                roomId: action.data.roomId
            };
        default:
            return state;
    }
};

export default socketReducer;
