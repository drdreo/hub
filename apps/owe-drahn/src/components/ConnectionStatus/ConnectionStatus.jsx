import { Signal, Unplug } from "lucide-react";
import React from "react";
import { useSelector } from "react-redux";
import { getConnectionManager } from "../../socket/ConnectionManager";
import "./ConnectionStatus.scss";

const ConnectionStatus = () => {
    const connectionStatus = useSelector(state => state.socket.connectionStatus);

    const getStatusDetails = () => {
        const connectionManager = getConnectionManager();
        const stats = connectionManager.getReconnectStats();
        const reconnectAttempts = stats.attempts;

        switch (connectionStatus) {
            case WebSocket.CONNECTING:
                return {
                    text:
                        reconnectAttempts > 0
                            ? `Reconnecting... (attempt ${reconnectAttempts})`
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
                // With new ConnectionManager, there's no hard max - it keeps trying
                const isReconnecting = reconnectAttempts > 0;
                return {
                    text: isReconnecting
                        ? `Reconnecting... (attempt ${reconnectAttempts})`
                        : "Disconnected",
                    class: "disconnected",
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
