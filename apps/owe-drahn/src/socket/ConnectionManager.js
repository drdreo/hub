/**
 * ConnectionManager
 *
 * Manages WebSocket lifecycle with:
 * - Exponential backoff with jitter
 * - Automatic reconnection (no hard limit)
 * - Health checks and heartbeat monitoring
 * - Mobile-aware connection handling
 * - Message queuing during offline periods
 */

class ConnectionManager {
    constructor(config = {}) {
        this.config = {
            url: config.url || this._getServerUrl(),
            baseDelay: config.baseDelay || 1000,
            maxDelay: config.maxDelay || 30000,
            healthCheckInterval: config.healthCheckInterval || 10000,
            maxMessageAge: config.maxMessageAge || 70000,
            jitterFactor: config.jitterFactor || 0.3,
            ...config
        };

        // Connection state
        this.socket = null;
        this.reconnectAttempts = 0;
        this.reconnectTimeout = null;
        this.healthCheckInterval = null;
        this.lastMessageTime = Date.now();
        this.isIntentionallyClosed = false;

        // Message queue
        this.messageQueue = [];
        this.isProcessingQueue = false;

        // Listeners
        this.listeners = {
            open: [],
            close: [],
            error: [],
            message: [],
            statusChange: []
        };

        // Flag to track if we've started
        this.hasStarted = false;

        // Bind event handlers
        this._setupBrowserEventHandlers();
    }

    /**
     * Dynamically determine server URL
     */
    _getServerUrl() {
        if (import.meta.env.VITE_DOMAIN && window.location.hostname === "localhost") {
            return import.meta.env.VITE_DOMAIN;
        }
        const protocol = window.location.protocol;
        const hostname = window.location.hostname;
        const port = "6969";
        return `${protocol}//${hostname}:${port}`;
    }

    /**
     * Register event listener
     */
    on(event, callback) {
        if (this.listeners[event]) {
            this.listeners[event].push(callback);
        }
        return () => this.off(event, callback);
    }

    /**
     * Unregister event listener
     */
    off(event, callback) {
        if (this.listeners[event]) {
            this.listeners[event] = this.listeners[event].filter(cb => cb !== callback);
        }
    }

    /**
     * Emit event to all listeners
     */
    _emit(event, ...args) {
        if (this.listeners[event]) {
            this.listeners[event].forEach(callback => {
                try {
                    callback(...args);
                } catch (error) {
                    console.error(`Error in ${event} listener:`, error);
                }
            });
        }
    }

    /**
     * Get current connection status
     */
    getStatus() {
        if (!this.socket) return WebSocket.CLOSED;
        return this.socket.readyState;
    }

    /**
     * Check if connection is open
     */
    isConnected() {
        return this.socket && this.socket.readyState === WebSocket.OPEN;
    }

    /**
     * Start the connection manager (call after middleware is initialized)
     */
    start() {
        if (this.hasStarted) {
            console.log("ConnectionManager already started");
            return;
        }
        this.hasStarted = true;
        this.connect();
    }

    /**
     * Connect to WebSocket server
     */
    connect() {
        // Clear any pending reconnection
        if (this.reconnectTimeout) {
            clearTimeout(this.reconnectTimeout);
            this.reconnectTimeout = null;
        }

        // Don't connect if already connecting or connected
        if (
            this.socket &&
            (this.socket.readyState === WebSocket.CONNECTING ||
                this.socket.readyState === WebSocket.OPEN)
        ) {
            console.log("Already connecting or connected");
            return;
        }

        this.isIntentionallyClosed = false;

        const wsUrl = this.config.url.replace("http", "ws") + "/ws?game=owedrahn";
        console.log("Connecting to WebSocket server at:", wsUrl);

        this.socket = new WebSocket(wsUrl);
        this._emit("statusChange", WebSocket.CONNECTING);

        this.socket.onopen = () => {
            console.log("WebSocket connected!");
            this.reconnectAttempts = 0;
            this.lastMessageTime = Date.now();

            this._emit("open");
            this._emit("statusChange", WebSocket.OPEN);

            this._startHealthCheck();
            this._processMessageQueue();
        };

        this.socket.onclose = event => {
            console.log("WebSocket disconnected!", event.code, event.reason);

            this._emit("close", event);
            this._emit("statusChange", WebSocket.CLOSED);

            this._stopHealthCheck();

            // Don't reconnect if intentionally closed
            if (this.isIntentionallyClosed) {
                console.log("Connection closed intentionally, not reconnecting");
                return;
            }

            // Attempt reconnection with exponential backoff + jitter
            this._scheduleReconnect();
        };

        this.socket.onerror = error => {
            console.error("WebSocket error:", error);
            this._emit("error", error);
        };

        this.socket.onmessage = event => {
            this.lastMessageTime = Date.now();
            this._emit("message", event);
        };
    }

    /**
     * Schedule reconnection with exponential backoff + jitter
     */
    _scheduleReconnect() {
        const baseDelay = Math.min(
            this.config.maxDelay,
            this.config.baseDelay * Math.pow(2, this.reconnectAttempts)
        );

        // Add jitter: random value between (1 - jitterFactor) and (1 + jitterFactor)
        const jitter = 1 + (Math.random() * 2 - 1) * this.config.jitterFactor;
        const delay = Math.floor(baseDelay * jitter);

        console.log(`Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts + 1})`);

        this.reconnectTimeout = setTimeout(() => {
            this.reconnectAttempts++;
            this.connect();
        }, delay);
    }

    /**
     * Disconnect from WebSocket server
     */
    disconnect() {
        this.isIntentionallyClosed = true;

        if (this.reconnectTimeout) {
            clearTimeout(this.reconnectTimeout);
            this.reconnectTimeout = null;
        }

        this._stopHealthCheck();

        if (this.socket) {
            this.socket.close();
            this.socket = null;
        }

        // Clear message queue on intentional disconnect
        this.messageQueue = [];
    }

    /**
     * Send message through WebSocket
     */
    send(type, data) {
        const message = { type, data };

        if (!this.isConnected()) {
            console.warn(`WebSocket not connected. Queuing message:`, type);
            this.messageQueue.push(message);
            return false;
        }

        try {
            this.socket.send(JSON.stringify(message));
            return true;
        } catch (error) {
            console.error("Failed to send message:", error);
            this.messageQueue.push(message);
            return false;
        }
    }

    /**
     * Process queued messages when connection is restored
     */
    _processMessageQueue() {
        if (this.isProcessingQueue || this.messageQueue.length === 0) {
            return;
        }

        if (!this.isConnected()) {
            console.warn("Cannot process queue - WebSocket not open");
            return;
        }

        this.isProcessingQueue = true;
        console.log(`Processing ${this.messageQueue.length} queued messages`);

        const messages = [...this.messageQueue];

        messages.forEach(message => {
            const success = this.send(message.type, message.data);
            if (!success) {
                console.error("Failed to send queued message:", message.type);
            } else {
                console.log("Sent queued message:", message.type);
            }
        });

        this.messageQueue = [];
        this.isProcessingQueue = false;
    }

    /**
     * Get number of queued messages
     */
    getQueueLength() {
        return this.messageQueue.length;
    }

    /**
     * Clear message queue
     */
    clearQueue() {
        if (this.messageQueue.length > 0) {
            console.log(`Clearing ${this.messageQueue.length} queued messages`);
            this.messageQueue = [];
        }
    }

    /**
     * Health check to detect dead connections
     */
    _startHealthCheck() {
        this._stopHealthCheck();

        this.healthCheckInterval = setInterval(() => {
            const timeSinceLastMessage = Date.now() - this.lastMessageTime;

            if (timeSinceLastMessage > this.config.maxMessageAge) {
                console.warn(
                    `No messages for ${Math.round(
                        timeSinceLastMessage / 1000
                    )}s, connection likely dead`
                );

                if (this.isConnected()) {
                    console.log("Forcing connection close to trigger reconnection");
                    this.socket.close();
                }
            }
        }, this.config.healthCheckInterval);
    }

    /**
     * Stop health check
     */
    _stopHealthCheck() {
        if (this.healthCheckInterval) {
            clearInterval(this.healthCheckInterval);
            this.healthCheckInterval = null;
        }
    }

    /**
     * Setup browser event handlers for visibility and network changes
     */
    _setupBrowserEventHandlers() {
        // Page Visibility API - handle backgrounding
        document.addEventListener("visibilitychange", () => {
            if (document.hidden) {
                console.log("Page hidden - connection may be suspended by browser");
            } else {
                console.log("Page visible - checking connection health");

                if (!this.isConnected()) {
                    console.log("Connection dead after returning from background, reconnecting...");
                    this.reconnectAttempts = 0; // Reset attempts on visibility change
                    this.connect();
                } else {
                    // Connection appears open, update check time
                    this.lastMessageTime = Date.now();
                }
            }
        });

        // Network change detection
        window.addEventListener("online", () => {
            console.log("Network back online");

            if (!this.isConnected()) {
                console.log("Reconnecting after network came online...");
                this.reconnectAttempts = 0; // Reset attempts for network recovery
                this.connect();
            }
        });

        window.addEventListener("offline", () => {
            console.log("Network offline - connection will be lost");
        });
    }

    /**
     * Get reconnection stats
     */
    getReconnectStats() {
        return {
            attempts: this.reconnectAttempts,
            queueLength: this.messageQueue.length,
            lastMessageTime: this.lastMessageTime,
            timeSinceLastMessage: Date.now() - this.lastMessageTime
        };
    }
}

// Singleton instance
let connectionManagerInstance = null;

/**
 * Get or create the singleton ConnectionManager instance
 */
export function getConnectionManager(config) {
    if (!connectionManagerInstance) {
        connectionManagerInstance = new ConnectionManager(config);
    }
    return connectionManagerInstance;
}

/**
 * Reset the singleton (useful for testing)
 */
export function resetConnectionManager() {
    if (connectionManagerInstance) {
        connectionManagerInstance.disconnect();
        connectionManagerInstance = null;
    }
}

export default ConnectionManager;
