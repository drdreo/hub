import { ExecutorContext, getPackageManagerCommand, logger } from "@nx/devkit";
import { exec } from "child_process";
import { promisify } from "util";
import { DeployDockerExecutorSchema } from "./schema";

const asyncExec = promisify(exec);

const runExecutor = async (options: DeployDockerExecutorSchema, context: ExecutorContext) => {
    try {
        // Default registry and image name if not provided
        const registry = options.registry || "ghcr.io";
        const username = process.env.GITHUB_ACTOR || "drdreo";

        const imageName = `${registry}/${username}/${context.projectName}`.toLowerCase();
        const imageTag = options.imageTag || "latest";
        const image = `${imageName}:${imageTag}`;

        const dockerFile = options.dockerFile || `apps/${context.projectName}/Dockerfile`;

        await loginToRegistry(registry, username);

        await buildDockerImage(dockerFile, image);

        await pushDockerImage(image);

        // Optional: Re-deploy service (customize as needed)
        if (options.redeploy) {
            await redeployService(options.serviceName);
        }

        return {
            success: true,
            message: `Deployed ${options.serviceName} successfully`
        };
    } catch (error) {
        console.error("Deployment failed:", error);

        return {
            success: false,
            message: error instanceof Error ? error.message : "Unknown error"
        };
    }
};

async function loginToRegistry(registry: string, username: string) {
    const token = process.env.GITHUB_TOKEN;

    if (!username || !token) {
        throw new Error("GitHub credentials not found");
    }

    await asyncExec(`echo ${token} | docker login ${registry} -u ${username} --password-stdin`);
}

async function buildDockerImage(dockerFile: string, tag: string) {
    const buildCommand = `docker build -t ${tag} -f ${dockerFile} .`;
    await asyncExec(buildCommand);
}

async function pushDockerImage(image: string) {
    await asyncExec(`docker push ${image}`);
}

async function redeployService(serviceName: string) {
    // ensure railway is installed
    await ensureRailwayInstalled();
    logger.log(`Redeploying service: ${serviceName}`);
    logger.log(process.env.RAILWAY_TOKEN);
    await asyncExec(`railway redeploy --service "${serviceName}" --yes`);
}

async function ensureRailwayInstalled() {
    try {
        const pm = getPackageManagerCommand();
        await asyncExec(`${pm.addDev} @railway/cli`);
        logger.log("Railway CLI installed successfully");
    } catch (error) {
        logger.error("Failed to install Railway CLI");
        logger.error(error);
        throw error;
    }
}

export default runExecutor;
