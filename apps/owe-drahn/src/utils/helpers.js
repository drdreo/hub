export const debounce = (func, delay) => {
    let inDebounce;
    return function (...args) {
        clearTimeout(inDebounce);
        inDebounce = setTimeout(() => func.apply(this, args), delay);
    };
};
