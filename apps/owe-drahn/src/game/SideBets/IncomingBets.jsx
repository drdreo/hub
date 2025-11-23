import { useDispatch } from "react-redux";
import { acceptSideBet, declineSideBet } from "../../socket/socket.actions";
import "./IncomingBets.scss";

const IncomingBets = ({ bets, players }) => {
    const dispatch = useDispatch();

    const getPlayerName = playerId => {
        const player = players.find(p => p.id === playerId);
        return player?.username || "Unknown";
    };

    const handleAccept = betId => {
        dispatch(acceptSideBet(betId));
    };

    const handleDecline = betId => {
        dispatch(declineSideBet(betId));
    };

    return (
        <div className="sidebet-section incoming">
            <h3>Incoming Challenges</h3>
            <div className="sidebet-list">
                {bets.map(bet => (
                    <div
                        key={bet.id}
                        className="sidebet-card incoming">
                        <div className="bet-info">
                            <p className="challenge-text">
                                <strong>{getPlayerName(bet.challengerId)}</strong> challenges you to a{" "}
                                <strong className="amount">${bet.amount}</strong> bet
                            </p>
                        </div>
                        <div className="bet-actions">
                            <button
                                className="button small danger"
                                onClick={() => handleDecline(bet.id)}>
                                Decline
                            </button>
                            <button
                                className="button small success"
                                onClick={() => handleAccept(bet.id)}>
                                Accept
                            </button>
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
};

export default IncomingBets;
