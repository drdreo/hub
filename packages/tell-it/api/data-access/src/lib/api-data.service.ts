import { Injectable, Logger } from "@nestjs/common";
import { InjectRepository } from "@nestjs/typeorm";
import { StoryData } from "@tell-it-shared/domain";
import { Repository, DeleteResult } from "typeorm";
import { StoryEntity } from "./entities/story.entity.ts";

@Injectable()
export class ApiDataService {
    private logger = new Logger(ApiDataService.name);

    constructor(@InjectRepository(StoryEntity) private storiesRepository: Repository<StoryEntity>) {}

    saveStories(stories: StoryData[]) {
        this.logger.log(`Saving ${stories.length} stories...`);

        const entities = this.storiesRepository.create(stories);
        return this.storiesRepository.save(entities);
    }

    findAllStories(): Promise<StoryEntity[]> {
        return this.storiesRepository.find();
    }

    findOneStory(id: number): Promise<StoryEntity | null> {
        return this.storiesRepository.findOneBy({ id });
    }

    remove(id: number): Promise<DeleteResult> {
        return this.storiesRepository.delete(id);
    }
}
