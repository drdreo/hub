/**
 * @typedef {Object} Player
 * @property {string} uid - Player's DB user id
 * @property {string} username - Player's name
 * @property {number} rank - Player's rank
 * @property {number} life - Number of lives remaining
 * @property {boolean} connected - Whether player is connected
 */

/**
 * @typedef {Object} DiceRoll
 * @property {number[]} dice - Array of dice values
 * @property {number} total - Total value of the roll
 */

/**
 * @typedef {Object} SideBet
 * @property {string} id - Side bet unique identifier
 * @property {string} challengerId - ID of player who challenged to bet
 * @property {string} challengerName - name of challenger
 * @property {string} opponentId - ID of player who received the proposal
 * @property {string} opponentName - name of opponent
 * @property {number} amount - Amount of the bet
 * @property {number} status - Status: 0 - 'pending', 1 - 'accepted', 2 -'declined', 3 -'resolved'
 */

/**
 * @typedef {Object} GameInfo
 * @property {string} message - Information message for the player
 */

/**
 * @typedef {Object} GameState
 * @property {DiceRoll | undefined} diceRoll - Latest server dice roll data
 * @property {number[] | undefined} rolledDice - Dice values for UI
 * @property {string} currentTurn - Current player's turn ID
 * @property {number} currentValue - Server value
 * @property {number} ui_currentValue - UI value
 * @property {Player[]} players - Server state of players
 * @property {Player[]} ui_players - UI state of players
 * @property {boolean} started - Whether game has started
 * @property {boolean} over - Whether game is over
 * @property {GameInfo} gameInfo - Game information messages
 * @property {SideBet[]} sideBets - Array of side bet objects
 */

/**
 * @typedef {Object} GameAction
 * @property {string} type - Action type
 * @property {*} [payload] - Action payload
 */

/** @type {GameState} */
const initialState = {
    diceRoll: undefined, // Latest server dice roll data
    rolledDice: undefined, // Dice values for UI
    currentTurn: "", // current players turn
    currentValue: 0, // Server value
    ui_currentValue: 0, // UI value
    mainBet: 1, // Main bet amount
    players: [],
    ui_players: [],
    started: false,
    over: false,
    gameInfo: { message: "" },
    sideBets: [] // Array of side bet objects
};

/**
 * Game reducer for managing Owe Drahn game state
 * @param {GameState} state - Current game state
 * @param {GameAction} action - Action to process
 * @returns {GameState} Updated game state
 */
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
                currentValue: action.payload.currentValue,
                mainBet: action.payload.mainBet ?? state.mainBet,
                sideBets: action.payload.sideBets ?? state.sideBets
            };
        case "GAME_OVER":
            return { ...state, over: true, started: false };
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

        case "SIDEBET_PROPOSED":
            return {
                ...state,
                sideBets: action.payload.bets || state.sideBets
            };
        case "SIDEBET_ACCEPTED":
            return {
                ...state,
                sideBets: action.payload.bets || state.sideBets
            };
        case "SIDEBET_DECLINED":
            return {
                ...state,
                sideBets: action.payload.bets || state.sideBets
            };
        case "SIDEBET_CANCELLED":
            return {
                ...state,
                sideBets: action.payload.bets || state.sideBets
            };

        default:
            return state;
    }
};

export default gameReducer;
