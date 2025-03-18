import type { TypeOrmModuleOptions } from "@nestjs/typeorm";

export function getProdConfig() {
    const dbUrl = process.env.DATABASE_URL;
    if (!dbUrl) {
        throw new Error("DATABASE_URL not set");
    }

    return {
        production: true,
        allowList: ["https://tell-it.pages.dev", "https://tell-it.drdreo.com"],
        typeOrm: {
            type: "postgres",
            url: dbUrl,
            autoLoadEntities: true,
            ssl: { rejectUnauthorized: false }
        } as TypeOrmModuleOptions
    };
}
