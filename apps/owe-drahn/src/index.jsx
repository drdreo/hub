import { applyMiddleware, compose } from "@reduxjs/toolkit";
import * as Sentry from "@sentry/react";
import { createBrowserHistory } from "history";
import { createRoot } from "react-dom/client";
import { Provider } from "react-redux";
import { BrowserRouter } from "react-router-dom";
import { legacy_createStore as createStore } from "redux";

import App from "./App.jsx";
import { FirebaseProvider } from "./auth/Firebase";
import { feedMiddleware } from "./game/Feed/feed.middleware";
import { createRootReducer } from "./reducers";
import * as serviceWorker from "./serviceWorker";
import { settingsMiddleware } from "./settings/settings.middleware";
import { getConnectionManager } from "./socket/ConnectionManager";
import { createSocketMiddleware } from "./socket/socket.middleware";

import "./index.css";

Sentry.init({
    dsn: import.meta.env.VITE_SENTRY_DSN,
    integrations: [Sentry.browserTracingIntegration()],
    tracesSampleRate: 0.1,
    // Set `tracePropagationTargets` to control for which URLs distributed tracing should be enabled
    tracePropagationTargets: [
        "localhost",
        /^wss:\/\/gameserver-production-23a9\.up\.railway\.app/,
        /^https:\/\/gameserver-production-23a9\.up\.railway\.app/
    ]
});

export const history = createBrowserHistory();

const composeEnhancers =
    typeof window === "object" && window.__REDUX_DEVTOOLS_EXTENSION_COMPOSE__
        ? window.__REDUX_DEVTOOLS_EXTENSION_COMPOSE__({})
        : compose;

// Create socket middleware and apply all middleware
const socketMiddleware = createSocketMiddleware();
const enhancer = composeEnhancers(
    applyMiddleware(settingsMiddleware, feedMiddleware, socketMiddleware)
);

const store = createStore(createRootReducer(history), enhancer);

// Start WebSocket connection after store is fully initialized
getConnectionManager().start();

const root = createRoot(document.getElementById("root"));
root.render(
    <Provider store={store}>
        <BrowserRouter>
            <FirebaseProvider>
                <App />
            </FirebaseProvider>
        </BrowserRouter>
    </Provider>
);

// If you want your app to work offline and load faster, you can change
// unregister() to register() below. Note this comes with some pitfalls.
// Learn more about service workers: https://bit.ly/CRA-PWA
serviceWorker.unregister();

export default store;
