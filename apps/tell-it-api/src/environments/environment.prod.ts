import type { TypeOrmModuleOptions } from "@nestjs/typeorm";
import { parse as parseConnectionString } from "pg-connection-string";

export function getProdConfig() {
    const dbUrl = process.env.DATABASE_URL;
    if (!dbUrl) {
        throw new Error("DATABASE_URL not set");
    }
    const connectionOptions = parseConnectionString(dbUrl);

    return {
        production: true,
        allowList: ["https://tell-it.pages.dev", "https://tell-it.drdreo.com"],
        typeOrm: {
            type: "postgres",
            host: connectionOptions.host,
            port: connectionOptions.port,
            user: connectionOptions.user,
            password: connectionOptions.password,
            database: connectionOptions.database,
            autoLoadEntities: true,
            ssl: { rejectUnauthorized: false }
        } as TypeOrmModuleOptions
    };
}
