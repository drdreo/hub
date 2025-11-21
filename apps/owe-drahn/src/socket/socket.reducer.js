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

// Helper to get session data with localStorage fallback
const getSessionData = (key) => {
    let value = sessionStorage.getItem(key);

    if (!value) {
        // Try localStorage backup (but only if recent - within 15 minutes)
        const timestamp = localStorage.getItem(`${key}_timestamp`);
        const age = Date.now() - parseInt(timestamp || "0");
        const FIFTEEN_MINUTES = 15 * 60 * 1000;

        if (age < FIFTEEN_MINUTES) {
            value = localStorage.getItem(`${key}_backup`);
            if (value) {
                // Restore to sessionStorage
                sessionStorage.setItem(key, value);
                console.log(`Restored ${key} from localStorage backup`);
            }
        } else if (timestamp) {
            // Clear old localStorage data
            localStorage.removeItem(`${key}_backup`);
            localStorage.removeItem(`${key}_timestamp`);
        }
    }

    return value;
};

const initialState = {
    socket: connectWebSocket(),
    clientId: getSessionData("clientId"),
    roomId: getSessionData("roomId"),
    connectionStatus: WebSocket.CLOSED
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
            // Clear session data when leaving room
            sessionStorage.removeItem("clientId");
            sessionStorage.removeItem("roomId");
            localStorage.removeItem("clientId_backup");
            localStorage.removeItem("roomId_backup");
            localStorage.removeItem("clientId_timestamp");
            return {
                ...state,
                clientId: null,
                roomId: null,
                reconnected: false
            };
        case CONNECTION_HANDSHAKE:
            sendMessage(socket, eventMap[CONNECTION_HANDSHAKE], {
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
