import { getSessionData } from "./session";
import { JOIN_ROOM_ERROR, JOINED_ROOM, RECONNECTED, RESET_RECONNECTED } from "./socket.actions";

const initialState = {
    clientId: getSessionData("clientId"),
    roomId: getSessionData("roomId"),
    connectionStatus: WebSocket.CONNECTING,
    reconnected: false,
    joinedRoom: false,
    joinError: null
};

/**
 * Pure socket reducer - only manages state, no side effects
 * All WebSocket communication is handled by the middleware
 */
const socketReducer = (state = initialState, action) => {
    switch (action.type) {
        case "CONNECTION_STATUS":
            return {
                ...state,
                connectionStatus: action.payload
            };

        case "GAME_LEAVE":
            return {
                ...state,
                clientId: null,
                roomId: null,
                reconnected: false,
                joinedRoom: false
            };

        case JOINED_ROOM:
            return {
                ...state,
                joinedRoom: true,
                clientId: action.data.clientId,
                roomId: action.data.roomId,
                joinError: null // Clear any previous errors
            };

        case JOIN_ROOM_ERROR:
            return {
                ...state,
                joinedRoom: false,
                joinError: action.error
            };

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
