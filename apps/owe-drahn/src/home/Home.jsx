import { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { useNavigate } from "react-router-dom"; // Use this hook for navigation
import { useFirebase } from "../auth/Firebase"; // Custom hook for Firebase context
import SignInGoogle from "../auth/SignIn/SignIn";
import { gameReset } from "../game/game.actions";
import { getRoomList, joinRoom } from "../socket/socket.actions";
import { getWebSocket } from "../socket/websocket";
import { debounce } from "../utils/helpers";

import "./Home.scss";

const Home = () => {
    const [room, setRoom] = useState("");
    const [username, setUsername] = useState("");
    const [usernameSetFromDB, setUsernameSetFromDB] = useState(false);
    const overview = useSelector(state => state.home.overview);
    const [formError, setFormError] = useState("");

    const navigate = useNavigate();
    const firebase = useFirebase();
    const authUser = useSelector(state => state.auth.authUser);
    const dispatch = useDispatch();

    useEffect(() => {
        console.log("Home mounted");
        sessionStorage.removeItem("playerId");
        fetchOverview();
        dispatch(gameReset());
    }, [dispatch]);

    useEffect(() => {
        if (authUser && authUser.username !== username && !usernameSetFromDB) {
            setUsername(authUser.username);
            setUsernameSetFromDB(true);
        } else if (!authUser) {
            setUsernameSetFromDB(false);
        }
    }, [authUser, usernameSetFromDB, username]);

    const updateRoom = room => {
        setRoom(room);
    };

    const updateUsername = evt => {
        const newUsername = evt.target.value;
        setUsername(newUsername);

        if (authUser) {
            updateDBUsername(newUsername);
        }
    };

    const updateDBUsername = debounce(username => {
        firebase.user(authUser.uid).update({ username });
    }, 200);

    const onRoomClick = (room, started) => {
        if (started) {
            navigate(`/game/${room}`);
        } else {
            updateRoom(room);
        }
    };

    const joinGame = () => {
        // Set up socket listener for join_room_result
        const socket = getWebSocket();

        // Create a one-time event listener for the join_room_result
        const messageHandler = event => {
            const messages = JSON.parse(event.data);
            messages.forEach(message => {
                if (message.type === "join_room_result") {
                    // Remove this listener since we only need it once
                    socket.removeEventListener("message", messageHandler);

                    if (message.success) {
                        const { clientId, roomId } = message.data;
                        sessionStorage.setItem("playerId", clientId);
                        navigate(`/game/${roomId}`);
                    } else {
                        let errMsg = message.error || "Failed to join game";
                        if (message.error?.includes("has started")) {
                            errMsg = `Game "${room}" has already started!`;
                        }
                        setFormError(errMsg);
                    }
                }
            });
        };

        socket.addEventListener("message", messageHandler);

        dispatch(joinRoom(room, username));
    };

    const fetchOverview = () => {
        // Request room list via WebSockets instead of HTTP
        dispatch(getRoomList());
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
                    onChange={evt => updateRoom(evt.target.value)}
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
