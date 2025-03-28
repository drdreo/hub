import { PlayerStats } from "./game.utils";

export type FormattedPlayer = {
    life: number;
    points: number;
    uid?: string;
    username: string;
    rank: number;
};

export class Player {
    isPlayersTurn = false;
    ready = false;
    life = 6;
    choosing = false;
    points = 0;
    connected = false;

    uid?: string; // only set if User is logged in
    rank = 0;

    get stats(): PlayerStats | undefined {
        return this._stats;
    }

    set stats(stats: PlayerStats) {
        this._stats = stats;
        this.rank = this.calculateRank(stats.totalGames);
    }

    private _stats?: PlayerStats; // only set if User is logged in

    constructor(readonly id: string, readonly username: string) {}

    getFormattedPlayer(): FormattedPlayer {
        return {
            life: this.life,
            points: this.points,
            uid: this.uid,
            username: this.username,
            rank: this.rank
        };
    }

    private calculateRank(totalGames: number): number {
        return Math.floor(totalGames / 10) + totalGames;
    }

    toJSON() {
        return {
            id: this.id,
            uid: this.uid,
            connected: this.connected,
            username: this.username,
            ready: this.ready,
            isPlayersTurn: this.isPlayersTurn,
            choosing: this.choosing,
            life: this.life,
            points: this.points,
            rank: this.rank,
            stats: this.stats
        };
    }
}
