import { Controller, HttpStatus, Get, Param, HttpException } from "@nestjs/common";
import { RoomService } from "@tell-it-api/game";
import type { HomeInfo, RoomResponse } from "@tell-it-shared/domain";

@Controller()
export class MainController {
    constructor(private readonly roomService: RoomService) {}

    @Get("/home")
    getHomeInfo(): HomeInfo {
        return { rooms: this.roomService.getAllRooms(), userCount: this.roomService.getUserCount() };
    }

    @Get("/room/:name")
    getTable(@Param("name") name: string): RoomResponse {
        const room = this.roomService.getRoom(name);
        if (room) {
            return {
                name: room.name,
                startTime: room.startTime!,
                users: room.getUsersPreview(),
                config: room.getConfig()
            };
        }

        throw new HttpException("Room does not exist!", HttpStatus.NOT_FOUND);
    }
}
