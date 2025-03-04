import { ExecutorContext } from "@nx/devkit";

import { DeployDockerExecutorSchema } from "./schema";
import executor from "./deploy-docker";

const options: DeployDockerExecutorSchema = {
    serviceName: "Test Service",
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

describe("DeployDocker Executor", () => {
    it("can run", async () => {
        const output = await executor(options, context);
        expect(output.success).toBe(true);
    });
});
