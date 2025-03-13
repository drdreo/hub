import { Logger } from "@nestjs/common";
import { ConfigService } from "@nestjs/config";
import { NestFactory } from "@nestjs/core";
import { AppModule } from "./app/app.module.ts";

async function bootstrap() {
    const app = await NestFactory.create(AppModule);
    const configService = app.get(ConfigService);
    const allowList = configService.get("allowList");
    console.log(allowList);
    app.enableCors({
        origin: (origin: string, callback: (err: Error | null, origin?: any) => void) => {
            console.log({ origin });
            // undefined if localhost
            if (allowList.indexOf(origin) !== -1 || !origin) {
                callback(null, { origin: true });
            } else {
                callback(new Error("Not allowed by CORS"), { origin: false });
            }
        }
    });
    const port = process.env.PORT || 3333;
    await app.listen(port);
    Logger.log(`🚀 Application is running on: http://localhost:${port}`, "main.ts");
}

bootstrap();
