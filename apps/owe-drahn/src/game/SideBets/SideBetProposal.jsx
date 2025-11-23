import { useState } from "react";
import { useDispatch } from "react-redux";
import { proposeSideBet } from "../../socket/socket.actions";
import "./SideBetProposal.scss";

const SideBetProposal = ({ opponent, onClose }) => {
    const dispatch = useDispatch();
    const [amount, setAmount] = useState("");
    const [error, setError] = useState("");

    const handleSubmit = e => {
        e.preventDefault();

        const betAmount = parseInt(amount, 10);
        if (isNaN(betAmount) || betAmount <= 0) {
            setError("Please enter a valid amount");
            return;
        }

        dispatch(proposeSideBet(opponent.id, betAmount));
        onClose();
    };

    return (
        <div
            className="sidebet-modal-overlay"
            onClick={onClose}>
            <div
                className="sidebet-modal"
                onClick={e => e.stopPropagation()}>
                <div className="sidebet-modal-header">
                    <h3>Propose Side Bet</h3>
                    <button
                        className="close-btn"
                        onClick={onClose}>
                        ×
                    </button>
                </div>

                <div className="sidebet-modal-content">
                    <p className="opponent-info">
                        Challenge <strong>{opponent.username}</strong> to a side bet
                    </p>

                    <form onSubmit={handleSubmit}>
                        <div className="form-group">
                            <label htmlFor="amount">Bet Amount</label>
                            <input
                                type="number"
                                id="amount"
                                value={amount}
                                onChange={e => {
                                    setAmount(e.target.value);
                                    setError("");
                                }}
                                placeholder="Enter amount"
                                min="1"
                                autoFocus
                            />
                            {error && <p className="error-message">{error}</p>}
                        </div>

                        <div className="modal-actions">
                            <button
                                type="button"
                                className="button light"
                                onClick={onClose}>
                                Cancel
                            </button>
                            <button
                                type="submit"
                                className="button">
                                Propose
                            </button>
                        </div>
                    </form>
                </div>
            </div>
        </div>
    );
};

export default SideBetProposal;
