import { Module } from "@nestjs/common";
import { TypeOrmModule } from "@nestjs/typeorm";
import { ApiDataService } from "./api-data.service.ts";
import { DataController } from "./data.controller.ts";
import { StoryEntity } from "./entities/story.entity.ts";

@Module({
    controllers: [DataController],
    providers: [ApiDataService],
    imports: [TypeOrmModule.forFeature([StoryEntity])],
    exports: [ApiDataService]
})
export class ApiDataAccessModule {}
