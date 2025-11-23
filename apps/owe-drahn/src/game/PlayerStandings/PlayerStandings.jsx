import { useSelector } from "react-redux";
import "./PlayerStandings.scss";

const PlayerStandings = () => {
    const players = useSelector(state => state.game.players);
    const enabled = useSelector(state => state.settings.standings.enabled);

    const hasBalance = players.every(player => player.balance != 0);
    if (!hasBalance || !enabled) {
        return null;
    }
    // clamp float numbers to 2 decimal places
    const balance = num => {
        return Math.round(num * 100) / 100;
    };
    return (
        <div className="standings">
            <ul>
                {players.map((player, idx) => (
                    <li key={idx}>
                        <span>{player.username}</span>
                        <span className={`balance ${balance(player.balance) > 0 ? "positive" : ""}`}>
                            {balance(player.balance)}
                        </span>
                    </li>
                ))}
            </ul>
        </div>
    );
};

export default PlayerStandings;
