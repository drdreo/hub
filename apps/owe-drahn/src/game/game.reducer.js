const initialState = {
    diceRoll: undefined, // Latest server dice roll data
    rolledDice: undefined, // Dice values for UI
    currentTurn: "", // current players turn
    currentValue: 0, // Server value
    ui_currentValue: 0, // UI value
    players: [],
    ui_players: [],
    started: false,
    over: false,
    error: undefined,
    gameInfo: { message: "" }
};

const gameReducer = (state = initialState, action) => {
    switch (action.type) {
        case "GAME_LEAVE":
            return { ...state, ...initialState };
        case "GAME_INIT":
            return {
                ...state,
                ...initialState,
                players: action.payload.players,
                ui_players: action.payload.players
            };
        case "GAME_STARTED":
            return { ...state, started: true, over: false };
        case "PATCH_UI_STATE":
            return {
                ...state,
                ui_currentValue: action.payload.currentValue
            };
        case "GAME_UPDATE":
            return {
                ...state,
                players: action.payload.players,
                ui_players: action.payload.players,
                started: action.payload.started,
                over: action.payload.over,
                currentTurn: action.payload.currentTurn,
                currentValue: action.payload.currentValue
            };
        case "GAME_OVER":
            return { ...state, over: true, started: false };
        case "GAME_ERROR":
            return {
                ...state,
                error: action.payload // Store the error in Redux state
            };
        case "PLAYER_UPDATE":
            if (action.payload.updateUI) {
                return {
                    ...state,
                    currentTurn: action.payload.currentTurn,
                    players: action.payload.players,
                    ui_players: action.payload.players
                };
            }
            return { ...state, players: action.payload.players };
        case "ROLLED_DICE":
            return {
                ...state,
                diceRoll: action.payload,
                currentValue: action.payload.total,
                gameInfo: { message: "" }
            };
        case "ANIMATED_DICE":
            return {
                ...state,
                rolledDice: action.payload.dice,
                ui_currentValue: action.payload.total,
                ui_players: state.players
            };
        case "PLAYER_LOST_LIFE": {
            const currentPlayerId = sessionStorage.getItem("clientId");
            const playersTurn = state.currentTurn === currentPlayerId;
            const message = playersTurn ? "Choose next Player or roll" : "";
            return {
                ...state,
                rolledDice: 0,
                ui_currentValue: 0,
                gameInfo: { message }
            };
        }

        case "PLAYER_CHOOSE_NEXT":
            return { ...state, gameInfo: { message: "" } };
        default:
            return state;
    }
};

export default gameReducer;
