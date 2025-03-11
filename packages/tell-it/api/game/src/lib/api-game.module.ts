import { Module } from "@nestjs/common";
import { ApiDataAccessModule } from "@tell-it-api/data-access";
import { RoomService } from "./room.service.js";

@Module({
    controllers: [],
    providers: [RoomService],
    imports: [ApiDataAccessModule],
    exports: [RoomService]
})
export class GameModule {}
