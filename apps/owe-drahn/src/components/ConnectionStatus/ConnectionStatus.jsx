import React from "react";
import { Signal, Unplug } from "lucide-react";
import { useSelector } from "react-redux";
import { getReconnectAttempts, getMaxReconnectAttempts } from "../../socket/websocket";
import "./ConnectionStatus.scss";

const ConnectionStatus = () => {
    const socket = useSelector(state => state.socket.socket);
    const readyState = socket?.readyState;

    const getStatusDetails = () => {
        const reconnectAttempts = getReconnectAttempts();
        const maxAttempts = getMaxReconnectAttempts();

        switch (readyState) {
            case WebSocket.CONNECTING:
                return {
                    text:
                        reconnectAttempts > 0
                            ? `Reconnecting... (${reconnectAttempts}/${maxAttempts})`
                            : "Connecting...",
                    class: "connecting",
                    icon: "•" // Simple dot
                };
            case WebSocket.OPEN:
                return {
                    text: "Connected",
                    class: "connected",
                    icon: <Signal size={15} />
                };
            case WebSocket.CLOSING:
                return {
                    text: "Closing connection...",
                    class: "closing",
                    icon: "•" // Simple dot
                };
            case WebSocket.CLOSED:
            default: {
                const isMaxedOut = reconnectAttempts >= maxAttempts;
                return {
                    text: isMaxedOut ? "Connection lost - Please refresh" : "Disconnected",
                    class: isMaxedOut ? "max-retries" : "disconnected",
                    icon: <Unplug size={15} />
                };
            }
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
