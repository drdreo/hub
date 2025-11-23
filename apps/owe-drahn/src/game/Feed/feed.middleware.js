import { feedMessage } from "./feed.actions";

export const feedMiddleware = store => next => action => {
    if (action.type === "GAME_OVER") {
        store.dispatch(feedMessage({ type: "GAME_OVER", winner: action.payload }));
    }

    if (action.type === "PLAYER_JOINED") {
        store.dispatch(feedMessage({ type: "PLAYER_JOINED", username: action.payload }));
    }

    if (action.type === "PLAYER_LEFT") {
        store.dispatch(feedMessage({ type: "PLAYER_LEFT", username: action.payload }));
    }

    if (action.type === "SIDEBET_PROPOSED") {
        const bets = action.payload.bets;
        const updatedBet = bets.find(bet => bet.id === action.payload.betId);
        if (updatedBet) {
            store.dispatch(
                feedMessage({
                    type: "SIDEBET_PROPOSED",
                    challenger: updatedBet.challengerName,
                    opponent: updatedBet.opponentName,
                    amount: updatedBet.amount
                })
            );
        }
    }

    if (action.type === "SIDEBET_ACCEPTED") {
        const bets = action.payload.bets;
        const updatedBet = bets.find(bet => bet.id === action.payload.betId);
        if (updatedBet) {
            store.dispatch(
                feedMessage({
                    type: "SIDEBET_ACCEPTED",
                    challenger: updatedBet.challengerName,
                    opponent: updatedBet.opponentName,
                    amount: updatedBet.amount
                })
            );
        }
    }

    if (action.type === "SIDEBET_DECLINED") {
        const bets = action.payload.bets;
        const updatedBet = bets.find(bet => bet.id === action.payload.betId);
        if (updatedBet) {
            store.dispatch(
                feedMessage({
                    type: "SIDEBET_DECLINED",
                    challenger: updatedBet.challengerName,
                    opponent: updatedBet.opponentName,
                    amount: updatedBet.amount
                })
            );
        }
    }

    return next(action);
};
