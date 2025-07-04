import { ExecutorContext } from "@nx/devkit";

import type { DeployDockerExecutorSchema } from "./schema.d.ts";
import executor from "./deploy-docker.ts";

const options: DeployDockerExecutorSchema = {
    dockerFile: "Dockerfile"
};
const context: ExecutorContext = {
    root: "",
    cwd: process.cwd(),
    isVerbose: false,
    projectGraph: {
        nodes: {},
        dependencies: {}
    },
    projectsConfigurations: {
        projects: {},
        version: 2
    },
    nxJsonConfiguration: {}
};

describe.skip("DeployDocker Executor", () => {
    it("can run", async () => {
        const output = await executor(options, context);
        expect(output.success).toBe(true);
    });
});
