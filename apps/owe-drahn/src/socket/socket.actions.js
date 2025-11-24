export const CONNECTION_HANDSHAKE = "CONNECTION_HANDSHAKE";
export const PLAYER_READY = "PLAYER_READY";
export const PLAYER_ROLL_DICE = "PLAYER_ROLL_DICE";
export const PLAYER_LOSE_LIFE = "PLAYER_LOSE_LIFE";
export const PLAYER_CHOOSE_NEXT = "PLAYER_CHOOSE_NEXT";
export const GET_ROOM_LIST = "GET_ROOM_LIST";
export const JOIN_ROOM = "JOIN_ROOM";
export const JOINED_ROOM = "JOINED_ROOM";
export const JOIN_ROOM_ERROR = "JOIN_ROOM_ERROR";
export const ROOM_ERROR = "ROOM_ERROR";
export const RECONNECT = "RECONNECT";
export const RESET_RECONNECTED = "RESET_RECONNECTED";
export const RECONNECTED = "RECONNECTED";
export const CONNECTION_STATUS = "CONNECTION_STATUS";
export const SIDEBET_PROPOSE = "SIDEBET_PROPOSE";
export const SIDEBET_ACCEPT = "SIDEBET_ACCEPT";
export const SIDEBET_DECLINE = "SIDEBET_DECLINE";
export const SIDEBET_CANCEL = "SIDEBET_CANCEL";
export const SET_MAIN_BET = "SET_MAIN_BET";

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
    RECONNECTED: "reconnect_result",
    SIDEBET_PROPOSE: "sidebet_propose",
    SIDEBET_ACCEPT: "sidebet_accept",
    SIDEBET_DECLINE: "sidebet_decline",
    SIDEBET_CANCEL: "sidebet_cancel",
    SET_MAIN_BET: "set_main_bet"
};

export const handshake = uid => {
    return {
        type: CONNECTION_HANDSHAKE,
        data: { uid }
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

export const joinRoomError = error => {
    return {
        type: JOIN_ROOM_ERROR,
        error
    };
};

export const roomError = error => {
    return {
        type: ROOM_ERROR,
        error
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

export const connectionStatus = status => {
    return {
        type: CONNECTION_STATUS,
        payload: status
    };
};

export const proposeSideBet = (opponentId, amount) => {
    return {
        type: SIDEBET_PROPOSE,
        data: { opponentId, amount }
    };
};

export const acceptSideBet = betId => {
    return {
        type: SIDEBET_ACCEPT,
        data: { betId }
    };
};

export const declineSideBet = betId => {
    return {
        type: SIDEBET_DECLINE,
        data: { betId }
    };
};

export const cancelSideBet = betId => {
    return {
        type: SIDEBET_CANCEL,
        data: { betId }
    };
};

export const setMainBet = amount => {
    return {
        type: SET_MAIN_BET,
        data: { amount }
    };
};
