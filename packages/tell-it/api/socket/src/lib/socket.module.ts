import { Module } from "@nestjs/common";
import { GameModule } from "@tell-it-api/game";
import { MainGateway } from "./main.gateway.ts";

@Module({
    providers: [MainGateway],
    imports: [GameModule],
    exports: []
})
export class SocketModule {}
