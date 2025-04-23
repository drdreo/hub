export const CONNECTION_HANDSHAKE = "CONNECTION_HANDSHAKE";
export const PLAYER_READY = "PLAYER_READY";
export const PLAYER_ROLL_DICE = "PLAYER_ROLL_DICE";
export const PLAYER_LOSE_LIFE = "PLAYER_LOSE_LIFE";
export const PLAYER_CHOOSE_NEXT = "PLAYER_CHOOSE_NEXT";
export const GET_ROOM_LIST = "GET_ROOM_LIST";
export const JOIN_ROOM = "JOIN_ROOM";
export const JOINED_ROOM = "JOINED_ROOM";
export const RECONNECT = "RECONNECT";
export const RESET_RECONNECTED = "RESET_RECONNECTED";
export const RECONNECTED = "RECONNECTED";

export const eventMap = {
    CONNECTION_HANDSHAKE: "handshake",
    PLAYER_READY: "ready",
    PLAYER_ROLL_DICE: "roll",
    PLAYER_LOSE_LIFE: "loseLife",
    PLAYER_CHOOSE_NEXT: "chooseNextPlayer",
    GET_ROOM_LIST: "get_room_list",
    JOIN_ROOM: "join_room",
    JOINED_ROOM: "join_room_result",
    RECONNECT: "reconnect",
    RECONNECTED: "reconnect_result"
};

export const handshake = (room, uid) => {
    return {
        type: CONNECTION_HANDSHAKE,
        data: { room, uid }
    };
};

export const ready = ready => {
    return {
        type: PLAYER_READY,
        data: ready
    };
};

export const rollDice = () => {
    return {
        type: PLAYER_ROLL_DICE
    };
};

export const loseLife = () => {
    return {
        type: PLAYER_LOSE_LIFE
    };
};

export const chooseNextPlayer = nextPlayerId => {
    return {
        type: PLAYER_CHOOSE_NEXT,
        data: { nextPlayerId }
    };
};

export const getRoomList = () => {
    return {
        type: GET_ROOM_LIST,
        data: { gameType: "owedrahn" }
    };
};

export const joinRoom = (roomId, playerName) => {
    return {
        type: JOIN_ROOM,
        data: { roomId, playerName, gameType: "owedrahn" }
    };
};

export const joinedRoom = ({ clientId, roomId }) => {
    return {
        type: JOINED_ROOM,
        data: { clientId, roomId }
    };
};

export const reconnect = (clientId, roomId) => {
    return {
        type: RECONNECT,
        data: { clientId, roomId }
    };
};

export const reconnected = ({ clientId, roomId }) => {
    return {
        type: RECONNECTED,
        data: { clientId, roomId }
    };
};

export const resetReconnected = () => {
    return {
        type: RESET_RECONNECTED
    };
};
