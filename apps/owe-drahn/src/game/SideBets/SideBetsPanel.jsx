import { useState } from "react";
import { useSelector } from "react-redux";
import ActiveBets from "./ActiveBets";
import IncomingBets from "./IncomingBets";
import PendingProposals from "./PendingProposals";
import SideBetProposal from "./SideBetProposal";
import "./SideBetsPanel.scss";

const SideBetsPanel = () => {
    const [isOpen, setIsOpen] = useState(false);
    const [showProposalForm, setShowProposalForm] = useState(false);
    const [selectedOpponent, setSelectedOpponent] = useState(null);

    const { sideBets, players, started } = useSelector(state => state.game);
    const clientId = useSelector(state => state.socket.clientId);

    // Only show during ready-up phase
    if (started) {
        return null;
    }

    // Filter bets by status
    const activeBets = sideBets.filter(bet => bet.status === 1);
    const pendingBets = sideBets.filter(bet => bet.status === 0 && bet.challengerId === clientId);
    const incomingBets = sideBets.filter(bet => bet.status === 0 && bet.opponentId === clientId);

    const handlePlayerSelect = player => {
        if (player.id !== clientId) {
            setSelectedOpponent(player);
            setShowProposalForm(true);
        }
    };

    const handleCloseProposal = () => {
        setShowProposalForm(false);
        setSelectedOpponent(null);
    };

    const togglePanel = () => {
        setIsOpen(!isOpen);
    };

    return (
        <>
            {/* Mobile: Floating Action Button */}
            <button
                className="sidebet-fab"
                onClick={togglePanel}
                aria-label="Side Bets">
                Side Bets{" "}
                {activeBets.length + pendingBets.length + incomingBets.length > 0 && (
                    <span className="badge">
                        {activeBets.length + pendingBets.length + incomingBets.length}
                    </span>
                )}
            </button>

            {/* Modal */}
            <div className={`sidebet-panel ${isOpen ? "open" : ""}`}>
                <div className="sidebet-panel-header">
                    <h2>Side Bets</h2>
                    <button
                        className="close-btn"
                        onClick={togglePanel}
                        aria-label="Close">
                        ×
                    </button>
                </div>

                <div className="sidebet-panel-content">
                    {/* Incoming bets (highest priority) */}
                    {incomingBets.length > 0 && (
                        <IncomingBets
                            bets={incomingBets}
                            players={players}
                        />
                    )}

                    {/* Active bets */}
                    {activeBets.length > 0 && (
                        <ActiveBets
                            bets={activeBets}
                            players={players}
                        />
                    )}

                    {/* Pending proposals */}
                    {pendingBets.length > 0 && (
                        <PendingProposals
                            bets={pendingBets}
                            players={players}
                        />
                    )}

                    {/* Proposal section */}
                    <div className="sidebet-section">
                        <h3>Propose Side Bet</h3>
                        <div className="player-select">
                            <p>Select an opponent:</p>
                            <div className="player-list">
                                {players
                                    .filter(p => p.id !== clientId)
                                    .map(player => (
                                        <button
                                            key={player.id}
                                            className="player-btn"
                                            onClick={() => handlePlayerSelect(player)}>
                                            {player.username}
                                        </button>
                                    ))}
                            </div>
                        </div>
                    </div>

                    {players.filter(p => p.id !== clientId).length === 0 && (
                        <p className="no-players">Waiting for other players to join...</p>
                    )}
                </div>
            </div>

            {/* Proposal Modal */}
            {showProposalForm && selectedOpponent && (
                <SideBetProposal
                    opponent={selectedOpponent}
                    onClose={handleCloseProposal}
                />
            )}
        </>
    );
};

export default SideBetsPanel;
