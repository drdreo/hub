import { useDispatch } from "react-redux";
import { cancelSideBet } from "../../socket/socket.actions";
import "./PendingProposals.scss";

const PendingProposals = ({ bets, players }) => {
    const dispatch = useDispatch();

    const getPlayerName = playerId => {
        const player = players.find(p => p.id === playerId);
        return player?.username || "Unknown";
    };

    const handleCancel = betId => {
        dispatch(cancelSideBet(betId));
    };

    return (
        <div className="sidebet-section">
            <h3>Your Proposals</h3>
            <div className="sidebet-list">
                {bets.map(bet => (
                    <div
                        key={bet.id}
                        className="sidebet-card pending">
                        <div className="bet-info">
                            <span className="opponent">to {getPlayerName(bet.opponentId)}</span>
                            <span className="amount">${bet.amount}</span>
                        </div>
                        <div className="bet-actions">
                            <span className="status-badge pending">Pending</span>
                            <button
                                className="button small danger"
                                onClick={() => handleCancel(bet.id)}>
                                Cancel
                            </button>
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
};

export default PendingProposals;
