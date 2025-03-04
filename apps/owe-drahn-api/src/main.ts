// IMPORTANT: Make sure to import `instrument.ts` at the top of your file.
import "./instrument";
// --------------
import { NestFactory } from "@nestjs/core";
import session from "express-session";
import { allowlist } from "./app/allow-list";
import { AppModule } from "./app/app.module";
import { EnvironmentService } from "./app/environment.service";

console.log("ENV: " + process.env.NODE_ENV);

async function bootstrap() {
    const app = await NestFactory.create(AppModule);
    app.enableCors({
        credentials: true,
        origin: (origin: string, callback: (...args: (object | null)[]) => void) => {
            console.log(origin);
            if (allowlist.indexOf(origin) !== -1 || !origin) {
                callback(null, { origin: true });
            } else {
                callback(new Error("Not allowed by CORS"), { origin: false });
            }
        }
    });

    app.use(
        session({
            secret: "secret",
            resave: false,
            saveUninitialized: true,
            cookie: {
                secure: true,
                httpOnly: true,
                maxAge: 1000 * 60 * 60 * 24
            }
        })
    );

    const envService = app.get<EnvironmentService>(EnvironmentService);
    await app.listen(envService.port);
    console.log("Server listening at port:" + envService.port);
}

bootstrap();
