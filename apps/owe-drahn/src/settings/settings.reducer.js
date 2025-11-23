import { TOGGLE_FEED, TOGGLE_SOUND, TOGGLE_STANDINGS, TOGGLE_SIDEBETS } from "./settings.actions";

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
    sideBets: {
        open: false
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
        case TOGGLE_SIDEBETS:
            return {
                ...state,
                sideBets: { ...state.sideBets, open: !state.sideBets.open }
            };
        default:
            return state;
    }
};

export default settingsReducer;
