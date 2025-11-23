import { useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { toggleSideBets } from "../../settings/settings.actions";
import ActiveBets from "./ActiveBets";
import IncomingBets from "./IncomingBets";
import PendingProposals from "./PendingProposals";
import SideBetProposal from "./SideBetProposal";
import "./SideBetsPanel.scss";

const SideBetsPanel = () => {
    const dispatch = useDispatch();
    const [showProposalForm, setShowProposalForm] = useState(false);
    const [selectedOpponent, setSelectedOpponent] = useState(null);

    const { sideBets, players, started } = useSelector(state => state.game);
    const clientId = useSelector(state => state.socket.clientId);
    const isOpen = useSelector(state => state.settings.sideBets?.open || false);

    // Only show FAB during ready-up phase
    const showFAB = !started;
    const showProposalSection = !started;

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
        dispatch(toggleSideBets());
    };

    return (
        <>
            {/* Floating Action Button - Only show during ready-up phase */}
            {showFAB && (
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
            )}

            {/* Backdrop */}
            <div
                className={`sidebet-backdrop ${isOpen ? "open" : ""}`}
                onClick={togglePanel}
            />

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

                    {/* Proposal section or game-started message */}
                    {showProposalSection ? (
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
                    ) : (
                        <div className="sidebet-section">
                            <p
                                style={{
                                    color: "rgba(255, 255, 255, 0.6)",
                                    textAlign: "center",
                                    padding: "20px",
                                    fontSize: "14px",
                                    fontStyle: "italic"
                                }}>
                                New side bets can only be placed before the game starts.
                            </p>
                        </div>
                    )}

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
