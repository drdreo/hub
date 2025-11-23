export const TOGGLE_FEED = "TOGGLE_FEED";
export const TOGGLE_SOUND = "TOGGLE_SOUND";
export const TOGGLE_STANDINGS = "TOGGLE_STANDINGS";
export const TOGGLE_SIDEBETS = "TOGGLE_SIDEBETS";

export const toggleFeed = () => {
    return {
        type: TOGGLE_FEED
    };
};

export const toggleSound = () => {
    return {
        type: TOGGLE_SOUND
    };
};

export const toggleStandings = () => {
    return {
        type: TOGGLE_STANDINGS
    };
};

export const toggleSideBets = () => {
    return {
        type: TOGGLE_SIDEBETS
    };
};
