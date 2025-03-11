import { Injectable, Logger } from "@nestjs/common";
import { WsException } from "@nestjs/websockets";
import { ApiDataService } from "@tell-it-api/data-access";
import { RoomInfo, StoryData } from "@tell-it-shared/domain";
import { Subject } from "rxjs";

import { RoomCommand, RoomCommandName, TellItRoom } from "./room/index.js";
import { User } from "./user/User.js";

const AUTO_DESTROY_DELAY = 30000; // after what time the room will be destroyed automatically, when no user is in it

@Injectable()
export class RoomService {
    rooms: TellItRoom[] = [];

    private _roomCommands$ = new Subject<RoomCommand>();
    roomCommands$ = this._roomCommands$.asObservable();

    private logger = new Logger(RoomService.name);
    private destroyTimeout: NodeJS.Timeout | undefined;

    constructor(private apiDataService: ApiDataService) {}

    userExists(userID: string) {
        return this.rooms.some(room => {
            return room.hasUser(userID);
        });
    }

    getRoom(roomName: string): TellItRoom | undefined {
        return this.rooms.find(room => room.name === roomName);
    }

    getAllRooms(): RoomInfo[] {
        return this.rooms
            .filter(room => room.getConfig().isPublic)
            .map(room => {
                return { name: room.name, started: room.hasStarted() };
            });
    }

    getUserCount(): number {
        return this.rooms.reduce((prev, room) => prev + room.getUserCount(), 0);
    }

    getRoomFinishVotes(roomName: string): string[] | undefined {
        return this.getRoom(roomName)?.getFinishVotes();
    }

    getStories(roomName: string): StoryData[] | undefined {
        return this.getRoom(roomName)?.getStories();
    }

    /**********************
     * HELPER METHODS
     **********************/

    sendCommand(command: RoomCommand) {
        this._roomCommands$.next(command);
    }

    createRoom(name: string): TellItRoom {
        const room = new TellItRoom(name, this._roomCommands$);
        room.commands$ = this._roomCommands$;
        this.rooms.push(room);
        return room;
    }

    start(room: string) {
        this.getRoom(room)?.start();
    }

    voteKick(roomName: string, userID: string, kickUserID: string) {
        const room = this.getRoom(roomName);
        if (room) {
            room.voteKick(userID, kickUserID);
        } else {
            throw new WsException(`Can not vote kick on Room[${roomName}] because it does not exist.`);
        }
    }

    createOrJoinRoom(roomName: string, userName: string): { userID: string } {
        let room = this.getRoom(roomName);

        if (!room) {
            this.logger.debug(`User[${userName}] created a room - ${roomName}!`);
            room = this.createRoom(roomName);
        }

        this.logger.debug(`User[${userName}] joining Room[${roomName}]!`);

        const userID = room.addUser(userName);

        return { userID };
    }

    userReconnected(userID: string): TellItRoom | undefined {
        for (const room of this.rooms) {
            const user = room.getUser(userID);
            if (user) {
                user.disconnected = false;
                return room;
            }
        }
        return undefined;
    }

    userLeft(roomName: string, userID: string): void {
        const user = this.getUser(userID);
        if (!user) {
            throw new WsException(`User[${userID}] not found!`);
        }
        const room = this.getRoom(roomName);
        if (!room) {
            throw new WsException(`Room[${roomName}] not found!`);
        }
        // if the room didnt start yet, just remove the player
        // functionality disabled. might add it back

        user.disconnected = true;

        // if every user disconnected, remove the table after some time
        if (this.destroyTimeout) {
            clearTimeout(this.destroyTimeout);
        }

        this.destroyTimeout = setTimeout(() => {
            if (room.isEmpty()) {
                room.destroy();
                this.removeRoom(roomName);
                this.sendCommand({ name: RoomCommandName.Info, room: room.name });
                this.logger.log(`Room[${room.name}] removed!`);
            }
        }, AUTO_DESTROY_DELAY);
    }

    submitNewText(roomName: string, userID: string, text: string) {
        const room = this.getRoom(roomName);
        if (room) {
            room.submitText(userID, text);
        } else {
            throw new WsException(
                `Can not submit a new text to Room[${roomName}] because it does not exist.`
            );
        }
    }

    voteFinish(roomName: string, userID: string) {
        const room = this.getRoom(roomName);
        if (room) {
            room.voteFinish(userID);
        } else {
            throw new WsException(
                `Can not vote finish on Room[${roomName}] because it does not exist.`
            );
        }
    }

    voteRestart(roomName: string, userID: string) {
        const room = this.getRoom(roomName);
        if (room) {
            room.voteRestart(userID);
        } else {
            throw new WsException(
                `Can not vote restart on Room[${roomName}] because it does not exist.`
            );
        }
    }

    async persistStories(stories: StoryData[]) {
        try {
            await this.apiDataService.saveStories(stories);
        } catch (e) {
            this.logger.error("Failed saving stories: " + e);
        }
    }

    private getUser(userID: string): User | undefined {
        for (const room of this.rooms) {
            const user = room.getUser(userID);
            if (user) {
                return user;
            }
        }
        return undefined;
    }

    private removeRoom(roomName: string): void {
        this.rooms = this.rooms.filter(r => r.name !== roomName);
    }
}
