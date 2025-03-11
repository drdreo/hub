import { parse as parseConnectionString } from "pg-connection-string";

const dbUrl = process.env.DATABASE_URL;
if (!dbUrl) {
    throw new Error("DATABASE_URL not set");
}
const connectionOptions = parseConnectionString(dbUrl);

export const environment = {
    production: true,
    allowList: ["https://tell-it.pages.dev", "https://tell-it.drdreo.com"],
    database: {
        host: connectionOptions.host,
        port: connectionOptions.port,
        user: connectionOptions.user,
        password: connectionOptions.password,
        database: connectionOptions.database
    }
};
