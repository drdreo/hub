import { useEffect, useState } from "react";

export const debounce = (func, delay) => {
    let inDebounce;
    return function (...args) {
        clearTimeout(inDebounce);
        inDebounce = setTimeout(() => func.apply(this, args), delay);
    };
};

export function useDebounce(value, delay) {
    const [debouncedValue, setDebouncedValue] = useState(value);

    useEffect(() => {
        // Set a timer to update the value after 'delay'
        const handler = setTimeout(() => {
            setDebouncedValue(value);
        }, delay);

        // Clean up the timer if 'value' changes before the delay expires
        return () => {
            clearTimeout(handler);
        };
    }, [value, delay]);

    return debouncedValue;
}
