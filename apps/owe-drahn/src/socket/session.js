export function setSessionData(key, value) {
    if (!value) {
        sessionStorage.removeItem(key);
        localStorage.removeItem(`${key}_backup`);
        localStorage.removeItem(`${key}_timestamp`);
        return;
    }

    sessionStorage.setItem(key, value);
    // Store in localStorage as fallback for mobile browsers
    localStorage.setItem(`${key}_backup`, value);
    localStorage.setItem(`${key}_timestamp`, Date.now().toString());
}

export const getSessionData = key => {
    let value = sessionStorage.getItem(key);

    if (!value) {
        // Try localStorage backup (but only if recent - within 15 minutes)
        const timestamp = localStorage.getItem(`${key}_timestamp`);
        const age = Date.now() - parseInt(timestamp || "0");
        const FIFTEEN_MINUTES = 15 * 60 * 1000;

        if (age < FIFTEEN_MINUTES) {
            value = localStorage.getItem(`${key}_backup`);
            if (value) {
                // Restore to sessionStorage
                sessionStorage.setItem(key, value);
                console.log(`Restored ${key} from localStorage backup`);
            }
        } else if (timestamp) {
            // Clear old localStorage data
            localStorage.removeItem(`${key}_backup`);
            localStorage.removeItem(`${key}_timestamp`);
        }
    }

    return value;
};
