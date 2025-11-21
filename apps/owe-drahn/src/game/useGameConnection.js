import { useEffect } from "react";
import { useDispatch, useSelector } from "react-redux";
import { handshake } from "../socket/socket.actions.js";

export const useGameConnection = () => {
    const dispatch = useDispatch();
    const authUser = useSelector(state => state.auth.authUser); // Redux hook for state

    useEffect(() => {
        // Perform auth user handshake, if not logged in, still handshake
        const uid = authUser?.uid;
        dispatch(handshake(uid));
    }, [authUser?.uid, dispatch]);
};
