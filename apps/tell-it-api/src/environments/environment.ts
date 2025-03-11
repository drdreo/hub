import type { TypeOrmModuleOptions } from "@nestjs/typeorm";

export function getDevConfig() {
    return {
        production: false,
        allowList: [
            "http://localhost:4200",
            "http://10.0.0.42:4200" // local IP
        ],
        typeOrm: {
            type: "sqlite",
            database: "tellit.sqlite",
            autoLoadEntities: true,
            synchronize: true
        } as TypeOrmModuleOptions
    };
}
