import { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { setMainBet } from "../../socket/socket.actions";
import "./MainBetInput.scss";
import { useDebounce } from "../../utils/helpers";

const MainBetInput = () => {
    const dispatch = useDispatch();
    const mainBetFromStore = useSelector(state => state.game.mainBet);
    const [localBet, setLocalBet] = useState(mainBetFromStore);
    const debouncedBet = useDebounce(localBet, 500);

    useEffect(() => {
        const val = parseFloat(debouncedBet);
        if (val > 0 && val !== mainBetFromStore) {
            dispatch(setMainBet(val));
        }
    }, [debouncedBet, mainBetFromStore, dispatch]);

    useEffect(() => {
        // Only update local if the store value is mathematically different.
        // This prevents "10." being forced to "10" while typing a decimal.
        if (parseFloat(localBet) !== mainBetFromStore) {
            setLocalBet(mainBetFromStore);
        }
    }, [mainBetFromStore]); // Remove localBet from dependency to avoid loops

    const handleMainBetChange = event => {
        setLocalBet(event.target.value);
    };

    return (
        <div className="main-bet-section">
            <label
                className="main-bet-label"
                htmlFor="main-bet-input">
                Main Bet:
                <input
                    id="main-bet-input"
                    type="number"
                    className="bet-input"
                    min="0.01"
                    step="any"
                    value={localBet}
                    onChange={handleMainBetChange}
                />
            </label>
        </div>
    );
};

export default MainBetInput;
