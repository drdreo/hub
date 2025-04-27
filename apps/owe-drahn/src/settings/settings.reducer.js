import { TOGGLE_FEED, TOGGLE_SOUND, TOGGLE_STANDINGS } from "./settings.actions";

const storedSettings = JSON.parse(localStorage.getItem("settings"));

const initialState = {
    feed: {
        enabled: true
    },
    sound: {
        enabled: true
    },
    standings: {
        enabled: true
    },
    ...storedSettings
};

const settingsReducer = (state = initialState, action) => {
    switch (action.type) {
        case TOGGLE_FEED:
            return {
                ...state,
                feed: { ...state.feed, enabled: !state.feed.enabled }
            };
        case TOGGLE_SOUND:
            return {
                ...state,
                sound: { ...state.sound, enabled: !state.sound.enabled }
            };
        case TOGGLE_STANDINGS:
            return {
                ...state,
                standings: { ...state.standings, enabled: !state.standings.enabled }
            };
        default:
            return state;
    }
};

export default settingsReducer;
