import { UserOverview } from "@tell-it-shared/domain";
import { Queue } from "@tell-it-shared/utils";
import { Story } from "../room/Story.ts";

export class User {
    disconnected = false;
    afk = false;
    kickVotes: Set<string> = new Set();

    storyQueue = new Queue<Story>();

    constructor(public id: string, public name: string) {}

    static getUserOverview(user: User): UserOverview {
        return {
            id: user.id,
            name: user.name,
            disconnected: user.disconnected,
            afk: user.afk,
            kickVotes: [...user.kickVotes],
            queuedStories: user.totalStories()
        };
    }

    reset() {
        this.afk = false;
        this.kickVotes = new Set();
        this.storyQueue = new Queue<Story>();
    }

    enqueueStory(story: Story) {
        this.storyQueue.enqueue(story);
    }

    getCurrentStory() {
        return this.storyQueue.peek();
    }

    dequeueCurrentStory() {
        return this.storyQueue.dequeue();
    }

    totalStories(): number {
        return this.storyQueue.size();
    }
}
