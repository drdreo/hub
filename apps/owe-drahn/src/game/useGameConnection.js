import { useEffect } from "react";
import { useDispatch, useSelector } from "react-redux";
import { handshake } from "../socket/socket.actions.js";

export const useGameConnection = room => {
    const dispatch = useDispatch();
    const authUser = useSelector(state => state.auth.authUser); // Redux hook for state
    const clientId = useSelector(state => state.socket.clientId);

    useEffect(() => {
        // Initialize socket listeners
        // Perform handshake
        const uid = authUser?.uid;
        dispatch(handshake(room, uid));
        // dispatch(reconnect(clientId, room));
    }, [room, authUser, clientId, dispatch]);
};
