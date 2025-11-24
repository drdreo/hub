import { useCallback, useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { useNavigate } from "react-router-dom"; // Use this hook for navigation
import { useFirebase } from "../auth/Firebase"; // Custom hook for Firebase context
import SignInGoogle from "../auth/SignIn/SignIn";
import { gameLeave } from "../game/game.actions";
import { getRoomList, joinRoom } from "../socket/socket.actions";
import { debounce } from "../utils/helpers";

import "./Home.scss";

const Home = () => {
    const [room, setRoom] = useState("");
    const [username, setUsername] = useState("");
    const [usernameSetFromDB, setUsernameSetFromDB] = useState(false);
    const overview = useSelector(state => state.home.overview);
    const [formError, setFormError] = useState("");
    const reconnected = useSelector(state => state.socket.reconnected);
    const joinedRoom = useSelector(state => state.socket.joinedRoom);
    const roomId = useSelector(state => state.socket.roomId);
    const connectionStatus = useSelector(state => state.socket.connectionStatus);
    const joinError = useSelector(state => state.socket.joinError);

    const navigate = useNavigate();
    const firebase = useFirebase();
    const authUser = useSelector(state => state.auth.authUser);
    const dispatch = useDispatch();

    // Request room list when connection is established
    useEffect(() => {
        if (connectionStatus === WebSocket.OPEN) {
            console.log("Connection ready, requesting room list");
            dispatch(getRoomList());
        }
    }, [connectionStatus, dispatch]);

    useEffect(() => {
        if (authUser && authUser.username !== username && !usernameSetFromDB) {
            setUsername(authUser.username);
            setUsernameSetFromDB(true);
        } else if (!authUser) {
            setUsernameSetFromDB(false);
        }
    }, [authUser, usernameSetFromDB, username]);

    // Handle reconnection - offer to rejoin existing game
    useEffect(() => {
        if (reconnected && roomId) {
            const shouldRejoin = window.confirm(
                `Game still in progress. Rejoin '${roomId}'? Cancel to leave room`
            );
            if (shouldRejoin) {
                navigate(`/game/${roomId}`);
            } else {
                dispatch(gameLeave());
            }
        }
    }, [reconnected, roomId, navigate, dispatch]);

    // Navigate to game when successfully joined
    useEffect(() => {
        if (joinedRoom && roomId) {
            console.log("Successfully joined room:", roomId);
            navigate(`/game/${roomId}`);
        }
    }, [joinedRoom, roomId, navigate]);

    useEffect(() => {
        if (joinError) {
            let errMsg = joinError;
            if (joinError.includes("has started")) {
                errMsg = `Game "${room}" has already started!`;
            }
            setFormError(errMsg);
        }
    }, [joinError, room]);

    const debouncedUpdateUsernameCallback = useCallback(
        debounce(username => {
            // Ensure authUser is available before trying to update the database
            if (authUser && firebase) {
                console.log("Saving username:", username);
                firebase.userUpdate(authUser.uid, { username });
            }
        }, 200),
        // If authUser or firebase change, we need a new function instance.
        [authUser, firebase]
    );

    const updateUsername = evt => {
        const newUsername = evt.target.value;
        setUsername(newUsername);

        if (authUser) {
            debouncedUpdateUsernameCallback(newUsername);
        }
    };

    const onRoomClick = (room, started) => {
        if (started) {
            navigate(`/game/${room}`);
        } else {
            setRoom(room);
        }
    };

    const joinGame = () => {
        setFormError(""); // Clear any previous errors
        dispatch(joinRoom(room, username));
    };

    return (
        <div className="page-container">
            <div className="overview">
                <div className="overview__total-players">
                    Online: <span>{overview.totalPlayers}</span>
                </div>
                Rooms
                <div className="overview__rooms">
                    {overview.rooms.map(room => (
                        <div
                            key={room.room}
                            className={`overview__rooms__entry ${room.started ? "has-started" : ""}`}
                            onClick={() => onRoomClick(room.room, room.started)}>
                            {room.started ? <span className="live"></span> : ""} {room.room}
                        </div>
                    ))}
                </div>
            </div>
            <h4>Owe Drahn</h4>
            <SignInGoogle className={`${authUser ? "is-hidden" : ""} sign-in-form`} />

            {authUser && (
                <>
                    <div>Hello {authUser.username}</div>
                    <button
                        className="link"
                        onClick={() => firebase.doSignOut()}>
                        Logout?
                    </button>
                </>
            )}
            <form
                className="form"
                onSubmit={e => {
                    e.preventDefault();
                    joinGame();
                }}>
                <input
                    className="input username"
                    value={username}
                    onChange={updateUsername}
                    placeholder="Username"
                />
                <input
                    className="input room"
                    value={room}
                    onChange={evt => setRoom(evt.target.value)}
                    placeholder="Room"
                />
                <button
                    className="button join"
                    disabled={!room}
                    type="submit">
                    Join
                </button>
            </form>

            <div className={`form__error ${!formError.length ? "is-invisible" : ""}`}>{formError}</div>
        </div>
    );
};

export default Home;
