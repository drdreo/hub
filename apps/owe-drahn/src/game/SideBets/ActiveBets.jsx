import "./ActiveBets.scss";

const ActiveBets = ({ bets, players }) => {
    const getPlayerName = playerId => {
        const player = players.find(p => p.id === playerId);
        return player?.username || "Unknown";
    };

    return (
        <div className="sidebet-section">
            <h3>Active</h3>
            <div className="sidebet-list">
                {bets.map(bet => (
                    <div
                        key={bet.id}
                        className="sidebet-card active">
                        <div className="bet-info">
                            <span className="opponent">
                                {getPlayerName(bet.challengerId)} vs. {getPlayerName(bet.opponentId)}
                            </span>
                            <span className="amount">${bet.amount}</span>
                        </div>
                        <div className="bet-status">
                            <span className="status-badge accepted">Active</span>
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
};

export default ActiveBets;
