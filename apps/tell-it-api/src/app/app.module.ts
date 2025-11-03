import { Logger, Module } from "@nestjs/common";
import { ConfigModule, ConfigService } from "@nestjs/config";
import { TypeOrmModule } from "@nestjs/typeorm";
import { ApiDataAccessModule } from "@tell-it-api/data-access";
import { GameModule } from "@tell-it-api/game";
import { SocketModule } from "@tell-it-api/socket";
import { getDevConfig } from "../environments/environment.ts";
import { getProdConfig } from "../environments/environment.prod.ts";
import { HealthController } from "./health.controller.ts";
import { MainController } from "./main.controller.ts";

const configuration = () => {
    if (process.env.NODE_ENV === "development") {
        Logger.log("Using dev config", "Config");
        return getDevConfig();
    }

    Logger.log("Using production config", "Config");
    return getProdConfig();
};

@Module({
    imports: [
        ConfigModule.forRoot({
            load: [configuration],
            isGlobal: true
        }),
        TypeOrmModule.forRootAsync({
            useFactory: (configService: ConfigService) => {
                return configService.getOrThrow("typeOrm");
            },
            inject: [ConfigService]
        }),
        ApiDataAccessModule,
        SocketModule,
        GameModule
    ],
    controllers: [MainController, HealthController],
    providers: []
})
export class AppModule {}
