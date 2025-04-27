import { useSelector } from "react-redux";
import "./PlayerStandings.scss";

const PlayerStandings = () => {
    const players = useSelector(state => state.game.players);
    const enabled = useSelector(state => state.settings.standings.enabled);

    const hasScore = players.every(player => player.score != 0);
    if (!hasScore || !enabled) {
        return null;
    }
    return (
        <div className="standings">
            <ul>
                {players.map((player, idx) => (
                    <li key={idx}>
                        <span>{player.username}</span>
                        <span className={`score ${player.score > 0 ? "positive" : ""}`}>
                            {player.score}
                        </span>
                    </li>
                ))}
            </ul>
        </div>
    );
};

export default PlayerStandings;
