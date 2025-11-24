import { Clock } from "lucide-react";
import { useSelector } from "react-redux";
import MainBetInput from "./MainBetInput";
import "./WaitingForPlayers.scss";

const WaitingForPlayers = () => {
    const players = useSelector(state => state.game.players);
    const readyCount = players.filter(p => p.ready).length;
    const totalPlayers = players.length;

    return (
        <div className="waiting-for-players">
            <div className="waiting-content">
                <Clock
                    className="waiting-icon"
                    size={48}
                    strokeWidth={1.5}
                />
                <p className="waiting-subtitle">Get ready to roll</p>
                <div className="ready-count">
                    <span className="count-text">
                        {readyCount} / {totalPlayers}
                    </span>
                    <span className="count-label">ready</span>
                </div>

                <MainBetInput />
            </div>
        </div>
    );
};

export default WaitingForPlayers;
