import { Controller, Get } from "@nestjs/common";
import { ApiDataService } from "./api-data.service.js";
import { StoryEntity } from "./entities/story.entity.js";

@Controller()
export class DataController {
    constructor(private apiDataService: ApiDataService) {}

    @Get("/stories")
    async getStories(): Promise<StoryEntity[]> {
        return await this.apiDataService.findAllStories();
    }
}
