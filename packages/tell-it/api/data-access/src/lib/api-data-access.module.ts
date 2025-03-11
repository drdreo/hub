import { Module } from "@nestjs/common";
import { TypeOrmModule } from "@nestjs/typeorm";
import { ApiDataService } from "./api-data.service.js";
import { DataController } from "./data.controller.js";
import { StoryEntity } from "./entities/story.entity.js";

@Module({
    controllers: [DataController],
    providers: [ApiDataService],
    imports: [TypeOrmModule.forFeature([StoryEntity])],
    exports: [ApiDataService]
})
export class ApiDataAccessModule {}
