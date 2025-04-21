import { useRef, useEffect } from "react";
import { useSelector } from "react-redux";

import FeedMessage from "./FeedMessage/FeedMessage";
import "./Feed.scss";

const Feed = () => {
    const feedRef = useRef(null);
    const messages = useSelector(state => state.feed.messages);
    const enabled = useSelector(state => state.settings.feed.enabled);

    const scrollToBottom = () => {
        if (feedRef.current) {
            feedRef.current.scrollTop = feedRef.current.scrollHeight + 21;
        }
    };

    useEffect(() => {
        if (feedRef.current) {
            const timeoutId = setTimeout(() => {
                scrollToBottom();
            }, 10);
            return () => clearTimeout(timeoutId);
        }
    }, [messages]); // Scroll to bottom when messages change

    if (!enabled) {
        return null;
    }

    return (
        <div
            className="feed"
            ref={feedRef}>
            {messages.map((message, index) => (
                <FeedMessage
                    message={message}
                    key={index}
                />
            ))}
        </div>
    );
};

export default Feed;
