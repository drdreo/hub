import { Module } from "@nestjs/common";
import { ConfigModule } from "@nestjs/config";
import { APP_FILTER } from "@nestjs/core";
import { SentryGlobalFilter, SentryModule } from "@sentry/nestjs/setup";
import { AppController } from "./app.controller";
import { AppService } from "./app.service";
import { DBService } from "./db/db.service";
import { EnvironmentService } from "./environment.service";
import { GameController } from "./game/game.controller";
import { GameService } from "./game/game.service";
import { SocketGateway } from "./game/socket/socket.gateway";
import { SocketService } from "./game/socket/socket.service";
import { UserController } from "./user/user.controller";

@Module({
    imports: [SentryModule.forRoot(), ConfigModule],
    controllers: [AppController, UserController, GameController],
    providers: [
        AppService,
        EnvironmentService,
        DBService,
        SocketGateway,
        SocketService,
        GameService,
        {
            provide: APP_FILTER,
            useClass: SentryGlobalFilter
        }
    ]
})
export class AppModule {}
