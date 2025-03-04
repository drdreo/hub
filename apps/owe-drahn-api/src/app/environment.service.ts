import { Injectable, Logger } from "@nestjs/common";
import * as path from "path";

export enum Environment {
    development = "development",
    production = "production",
    testing = "testing"
}

@Injectable()
export class EnvironmentService {
    readonly credentialsDir =
        this.env === "production"
            ? path.join(__dirname, "../src/credentials")
            : path.join(__dirname, "../src/credentials");

    private readonly _env = process.env.NODE_ENV || Environment.development;
    private readonly _port = process.env.PORT || 4000;

    private logger = new Logger(EnvironmentService.name);

    constructor() {
        this.logger.log("EnvironmentService - Constructed!");
    }

    get env(): string {
        return this._env;
    }

    get port(): number {
        return Number(this._port);
    }
}
