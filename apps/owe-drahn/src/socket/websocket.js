const SERVER_URL = import.meta.env.VITE_DOMAIN;

let socket;

export const connectWebSocket = () => {
    socket = new WebSocket(SERVER_URL.replace("http", "ws") + "/ws?game=owedrahn"); //http to ws/wss

    socket.onopen = () => {
        console.log("WebSocket connected!");
    };

    socket.onclose = () => {
        console.log("WebSocket disconnected!");
        // Attempt to reconnect after a delay
        setTimeout(connectWebSocket, 3000);
    };

    socket.onerror = error => {
        console.error("WebSocket error:", error);
    };

    return socket;
};

export const getWebSocket = () => {
    return socket;
}