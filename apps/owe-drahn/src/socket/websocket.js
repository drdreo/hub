const SERVER_URL = import.meta.env.VITE_DOMAIN;

let socket;
let reconnectAttempts = 0;
let lastMessageTime = Date.now();
let healthCheckInterval = null;
let reconnectTimeout = null;

const MAX_RECONNECT_ATTEMPTS = 10;
const BASE_DELAY = 1000; // 1 second
const HEALTH_CHECK_INTERVAL = 10000; // 10 seconds
const MAX_MESSAGE_AGE = 70000; // 70 seconds (longer than server ping of 54s)

export const connectWebSocket = () => {
    // Clear any pending reconnection timeout
    if (reconnectTimeout) {
        clearTimeout(reconnectTimeout);
        reconnectTimeout = null;
    }

    const wsUrl = SERVER_URL.replace("http", "ws") + "/ws?game=owedrahn"; //http to ws/wss
    console.log("Connecting to WebSocket server at:", wsUrl);

    socket = new WebSocket(wsUrl);

    socket.onopen = () => {
        console.log("WebSocket connected!");
        reconnectAttempts = 0; // Reset on successful connection
        lastMessageTime = Date.now();
        
        // Start health check monitoring
        startHealthCheck();
    };

    socket.onclose = (event) => {
        console.log("WebSocket disconnected!", event.code, event.reason);
        
        // Stop health check
        stopHealthCheck();
        
        // Attempt reconnection with exponential backoff
        if (reconnectAttempts < MAX_RECONNECT_ATTEMPTS) {
            const delay = Math.min(30000, BASE_DELAY * Math.pow(2, reconnectAttempts));
            console.log(`Reconnecting in ${delay}ms (attempt ${reconnectAttempts + 1}/${MAX_RECONNECT_ATTEMPTS})`);
            
            reconnectTimeout = setTimeout(() => {
                reconnectAttempts++;
                connectWebSocket();
            }, delay);
        } else {
            console.error('Max reconnection attempts reached. Please refresh the page.');
        }
    };

    socket.onerror = error => {
        console.error("WebSocket error:", error);
    };

    return socket;
};

// Track when messages are received to monitor connection health
export const updateLastMessageTime = () => {
    lastMessageTime = Date.now();
};

// Health check to detect dead connections
const startHealthCheck = () => {
    stopHealthCheck(); // Clear any existing interval
    
    healthCheckInterval = setInterval(() => {
        const timeSinceLastMessage = Date.now() - lastMessageTime;
        
        if (timeSinceLastMessage > MAX_MESSAGE_AGE) {
            console.warn(`No messages for ${Math.round(timeSinceLastMessage / 1000)}s, connection likely dead`);
            
            // Force close to trigger reconnection if socket thinks it's open
            if (socket && socket.readyState === WebSocket.OPEN) {
                console.log('Forcing connection close to trigger reconnection');
                socket.close();
            }
        }
    }, HEALTH_CHECK_INTERVAL);
};

const stopHealthCheck = () => {
    if (healthCheckInterval) {
        clearInterval(healthCheckInterval);
        healthCheckInterval = null;
    }
};

export const getWebSocket = () => {
    return socket;
};

export const getReconnectAttempts = () => {
    return reconnectAttempts;
};

export const getMaxReconnectAttempts = () => {
    return MAX_RECONNECT_ATTEMPTS;
};

// Page Visibility API - handle backgrounding
document.addEventListener('visibilitychange', () => {
    if (document.hidden) {
        console.log('Page hidden - connection may be suspended by browser');
    } else {
        console.log('Page visible - checking connection health');
        const currentSocket = getWebSocket();
        
        // Check if connection is actually alive
        if (!currentSocket || currentSocket.readyState !== WebSocket.OPEN) {
            console.log('Connection dead after returning from background, reconnecting...');
            connectWebSocket();
        } else {
            // Connection appears open, but might be stale - update check time
            lastMessageTime = Date.now();
        }
    }
});

// Network change detection
window.addEventListener('online', () => {
    console.log('Network back online');
    const currentSocket = getWebSocket();
    
    if (!currentSocket || currentSocket.readyState !== WebSocket.OPEN) {
        console.log('Reconnecting after network came online...');
        reconnectAttempts = 0; // Reset attempts for network recovery
        connectWebSocket();
    }
});

window.addEventListener('offline', () => {
    console.log('Network offline - connection will be lost');
});
