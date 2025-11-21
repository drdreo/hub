import diceRoller from "dice-roller-3d";
import { Howl } from "howler";
import { useEffect, useRef, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { useNavigate, useParams } from "react-router-dom";
import yourTurnAudio from "../assets/sounds/your_turn.mp3";
import Settings from "../settings/Settings";

import {
    chooseNextPlayer,
    loseLife,
    ready,
    resetReconnected,
    rollDice
} from "../socket/socket.actions";
import { gameLeave } from "./game.actions";
import Feed from "./Feed/Feed";
import { feedMessage } from "./Feed/feed.actions";

import "./Game.scss";
import { animatedDice, patchUIState } from "./game.actions";
import GameInfo from "./GameInfo/GameInfo";
import LifeLoseBtn from "./LifeLoseBtn/LifeLoseBtn";

import Player from "./Player/Player";
import PlayerStandings from "./PlayerStandings/PlayerStandings";
import RollButton from "./RollButton/RollButton";
import RolledDice from "./RolledDice/RolledDice.jsx";
import { useGameConnection } from "./useGameConnection.js";

const MIN_VAL_TO_OWE_DRAHN = 10;

const Game = () => {
    const dispatch = useDispatch();
    const navigate = useNavigate();
    const { room } = useParams();
    const clientId = useSelector(state => state.socket.clientId);
    useGameConnection(room);

    const settings = useSelector(state => state.settings);
    const { diceRoll, currentValue, currentTurn, ui_currentValue, ui_players, players, started, over } =
        useSelector(state => state.game);
    const reconnected = useSelector(state => state.socket.reconnected);
    const roomError = useSelector(state => state.socket.roomError);

    const [animatingDice, setAnimatingDice] = useState(false);
    const [animatingHeart, setAnimatingHeart] = useState(false);
    const [isRolling, setIsRolling] = useState(false);
    const diceRef = useRef(null);
    const sfx = {
        yourTurn: {
            played: false,
            audio: new Howl({ src: [yourTurnAudio] })
        }
    };

    useEffect(() => {
        if (reconnected) {
            dispatch(patchUIState({ currentValue }));
            dispatch(resetReconnected());
        }
    }, [reconnected, currentValue]);

    // Handle room errors - redirect to home if room no longer exists
    useEffect(() => {
        if (roomError) {
            console.error("Room error detected:", roomError);

            // Show user-friendly message
            const errorMessage =
                roomError.includes("not found") || roomError.includes("does not exist")
                    ? "This room no longer exists."
                    : "There was a problem with the room.";

            alert(`${errorMessage}\n\nYou will be redirected to the lobby.`);

            // Clean up and go home
            dispatch(gameLeave());
            navigate("/");
        }
    }, [roomError, dispatch, navigate]);

    // Detect if room became empty or invalid
    useEffect(() => {
        // If game started but all players left (shouldn't happen normally)
        if (started && players.length === 0) {
            console.warn("Game started but no players remain - room may be invalid");

            // Give a grace period in case it's a temporary state during reconnection
            const timeout = setTimeout(() => {
                if (players.length === 0) {
                    alert("All players have left the game.\n\nYou will be redirected to the lobby.");
                    dispatch(gameLeave());
                    navigate("/");
                }
            }, 3000); // 3 second grace period

            return () => clearTimeout(timeout);
        }
    }, [started, players.length, dispatch, navigate]);

    const getPlayer = () => {
        return players.find(player => player.id === clientId);
    };

    const player = getPlayer();
    const isPlayersTurn = player?.id === currentTurn;
    const isChoosing = isPlayersTurn && player?.choosing;

    useEffect(() => {
        if (!diceRoll || animatingDice) return;

        animateDice(diceRoll.dice, diceRoll.total).then(() => {
            let msg;
            if (diceRoll.total > 15) {
                msg = {
                    type: "LOST",
                    username: diceRoll.player.username,
                    dice: diceRoll.dice,
                    total: diceRoll.total
                };
            } else if (!over) {
                msg = {
                    type: "ROLLED_DICE",
                    username: diceRoll.player.username,
                    dice: diceRoll.dice,
                    total: diceRoll.total
                };
            }

            if (msg) {
                dispatch(feedMessage(msg));
            }
        });
    }, [diceRoll]);

    useEffect(() => {
        // Change global volume.
        // Howler.mute(!settings.sound.enabled);

        if (!player || !settings.sound.enabled) return;

        if (!animatingDice) {
            // If it's the player's turn and the sound hasn't played yet, play it
            if (isPlayersTurn && !isChoosing && !sfx.yourTurn.played) {
                sfx.yourTurn.played = true;
                sfx.yourTurn.audio.play();
            }

            // If it's not the player's turn, reset the sound played flag
            if (!isPlayersTurn && sfx.yourTurn.played) {
                sfx.yourTurn.played = false;
            }
        }
    }, [isPlayersTurn, isChoosing, animatingDice, settings.sound.enabled]);

    const handleReady = () => {
        const isReady = !player.ready;
        dispatch(ready(isReady));
    };

    const handleRollDice = () => {
        if (isPlayersTurn && !animatingDice) {
            if (!isRolling) {
                setIsRolling(true);
                setTimeout(() => {
                    setIsRolling(false);
                }, 1300);
            }
            dispatch(rollDice());
        }
    };

    const handleLoseLife = () => {
        const player = getPlayer();
        if (isPlayersTurn && player.life > 1 && currentValue >= MIN_VAL_TO_OWE_DRAHN) {
            if (!animatingHeart) {
                setAnimatingHeart(true);
                // remove the animation class after some arbitrary time. Player won't trigger this again soon
                setTimeout(() => {
                    setAnimatingHeart(false);
                }, 2500);
            }
            dispatch(loseLife());
        }
    };

    const handleChooseNextPlayer = playerId => {
        const player = getPlayer();
        if (isPlayersTurn && player.choosing) {
            dispatch(chooseNextPlayer(playerId));
        }
    };

    const animateDice = (dice, total) => {
        setAnimatingDice(true);

        return new Promise(resolve => {
            diceRoller({
                element: diceRef.current,
                numberOfDice: 1,
                delay: 1250,
                callback: () => {
                    setAnimatingDice(false);
                    dispatch(animatedDice({ dice, total }));
                    resolve();
                },
                values: [dice],
                noSound: !settings.sound.enabled
            });
        });
    };

    const getPlayerPosition = (index, totalPlayers) => {
        const vw = Math.min(window.innerWidth, window.innerHeight);
        const radius = vw < 800 ? vw * 0.35 : 250; // 30% of viewport on mobile, fixed on desktop
        const degrees = (360 / totalPlayers) * index;
        return {
            transform: `
            translateX(-50%)
            translateY(-50%)
            rotate(${degrees}deg) 
            translateY(-${radius}px) 
            rotate(-${degrees}deg)
            `
        };
    };

    // maybe is spectator
    let controls;
    if (player) {
        let controlButton;

        if (!over || animatingDice) {
            if (!animatingDice) {
                if (players.length === 1) {
                    controlButton = "Waiting for Players";
                } else {
                    controlButton = (
                        <button
                            className={`button ${player.ready ? "success" : "primary"}`}
                            onClick={() => handleReady()}>
                            Ready
                        </button>
                    );
                }
            }

            if (started || animatingDice) {
                const isWaiting = !isPlayersTurn || animatingDice;

                controlButton = (
                    <div
                        style={{ display: "flex" }}
                        className={`${isWaiting ? "waiting" : ""}`}>
                        <RollButton
                            rolling={isRolling}
                            disabled={isWaiting}
                            onClick={handleRollDice}
                        />
                        <LifeLoseBtn
                            animating={animatingHeart}
                            disabled={
                                isWaiting || player.life <= 1 || ui_currentValue < MIN_VAL_TO_OWE_DRAHN
                            }
                            onClick={handleLoseLife}
                        />
                    </div>
                );
            }
            controls = <div className="controls">{controlButton}</div>;
        }
    }

    return (
        <div className="page-container">
            <RolledDice />

            {controls}

            <div className="players-list">
                {ui_players.map((player, index) => (
                    <Player
                        player={player}
                        started={started}
                        connected={player.connected}
                        isPlayersTurn={player.id === currentTurn}
                        choosing={isChoosing}
                        key={player.id}
                        style={getPlayerPosition(index, players.length)}
                        onClick={() => handleChooseNextPlayer(player.id)}
                    />
                ))}
            </div>

            <div
                className="dice"
                ref={diceRef}
            />
            <Feed />
            <Settings className="settings" />
            <GameInfo />
            <PlayerStandings />
        </div>
    );
};

export default Game;
