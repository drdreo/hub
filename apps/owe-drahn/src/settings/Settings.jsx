import {
    ListOrdered,
    LogOut,
    MessageSquareOff,
    MessageSquareText,
    Volume2,
    VolumeOff,
    X
} from "lucide-react";
import { useState } from "react";
import { useDispatch, useSelector } from "react-redux";

import "./Settings.scss";
import { useNavigate } from "react-router-dom";
import { gameLeave } from "../game/game.actions";
import { toggleFeed, toggleSound, toggleStandings } from "./settings.actions";

const Speaker = ({ disabled }) => (disabled ? <VolumeOff /> : <Volume2 />);
const Feed = ({ disabled }) => (disabled ? <MessageSquareOff /> : <MessageSquareText />);
const Standing = ({ disabled }) => (
    <div style={{ position: "relative" }}>
        <ListOrdered />
        {disabled && (
            <X
                style={{
                    position: "absolute",
                    top: "50%",
                    left: "50%",
                    transform: "translate(-50%, -50%)"
                }}
            />
        )}
    </div>
);

const Settings = props => {
    const dispatch = useDispatch();
    const navigate = useNavigate();
    const [open, setOpen] = useState(false);
    const settings = useSelector(state => state.settings);
    const { feed, sound, standings } = settings;
    const soundEnabled = sound.enabled;
    const feedEnabled = feed.enabled;
    const standingsEnabled = standings.enabled;
    const menuClass = !open ? " menu--closed " : "menu--open";

    const toggleMenu = () => {
        setOpen(!open);
    };
    const handleToggleSound = () => dispatch(toggleSound());
    const handleToggleFeed = () => dispatch(toggleFeed());
    const handleToggleStandings = () => dispatch(toggleStandings());

    const handleLeaveRoom = () => {
        const shouldLeave = window.confirm("Are you sure you want to leave this game?");

        if (shouldLeave) {
            // Dispatch leave action to clean up room state and notify server
            dispatch(gameLeave());
            // Navigate to home after leaving
            navigate("/");
        }
    };

    return (
        <div className={`menu ${menuClass} ${props.className}`}>
            <div
                className={`hamburger`}
                onClick={toggleMenu}>
                <div className="hamburger-box">
                    <div className="hamburger-inner" />
                </div>
            </div>

            <div className={`menu-entries`}>
                <button
                    className="menu__button"
                    onClick={handleToggleSound}>
                    <Speaker disabled={!soundEnabled} />
                </button>
                <button
                    className="menu__button"
                    onClick={handleToggleFeed}>
                    <Feed disabled={!feedEnabled} />
                </button>
                <button
                    className="menu__button"
                    onClick={handleToggleStandings}>
                    <Standing disabled={!standingsEnabled} />
                </button>
                <button
                    className="menu__button"
                    onClick={handleLeaveRoom}
                    title="Leave Game">
                    <LogOut />
                </button>
            </div>
        </div>
    );
};

export default Settings;
