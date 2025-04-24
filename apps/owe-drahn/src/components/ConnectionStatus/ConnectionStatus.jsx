import React from "react";
import { useSelector } from "react-redux";
import "./ConnectionStatus.scss";

const ConnectionStatus = () => {
    const socket = useSelector(state => state.socket.socket);
    const connectionStatus = useSelector(state => state.socket.connectionStatus);
    const readyState = socket?.readyState;
    console.log({ readyState, connectionStatus});
    const getStatusDetails = () => {
        switch (readyState) {
            case WebSocket.CONNECTING:
                return {
                    text: "Connecting...",
                    class: "connecting",
                    icon: "•" // Simple dot
                };
            case WebSocket.OPEN:
                return {
                    text: "Connected",
                    class: "connected",
                    icon: "•" // Simple dot
                };
            case WebSocket.CLOSING:
                return {
                    text: "Closing connection...",
                    class: "closing",
                    icon: "•" // Simple dot
                };
            case WebSocket.CLOSED:
            default:
                return {
                    text: "Disconnected",
                    class: "disconnected",
                    icon: "•" // Simple dot
                };
        }
    };

    const statusDetails = getStatusDetails();

    return (
        <div className="connection-status-container">
            <div
                className={`connection-status ${statusDetails.class}`}
                title={statusDetails.text}>
                <span className="status-icon">{statusDetails.icon}</span>
            </div>
            <div className="connection-tooltip">{statusDetails.text}</div>
        </div>
    );
};

export default ConnectionStatus;
