import { ExecutorContext, getPackageManagerCommand, logger } from "@nx/devkit";
import { default as axios } from 'axios';
import { exec } from "child_process";
import { promisify } from "util";
import type { DeployDockerExecutorSchema } from "./schema.d.ts";

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
            await redeployService(options.projectId);
        }

        return {
            success: true,
            message: `Deployed ${options.projectId} successfully`
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

    logger.log(`Logging in to ${registry} as ${username}`);
    await asyncExec(`echo ${token} | docker login ${registry} -u ${username} --password-stdin`);
}

async function buildDockerImage(dockerFile: string, tag: string) {
    logger.log(`Building docker image: ${dockerFile} -t ${tag}`);
    const buildCommand = `docker build -t ${tag} -f ${dockerFile} .`;
    await asyncExec(buildCommand);
}

async function pushDockerImage(image: string) {
    logger.log(`Pushing docker image: ${image}`);
    await asyncExec(`docker push ${image}`);
}

async function getLastDeployment(projectId: string): Promise<any> {
    const gql = `
    query lastDeployment($projectId: String!) {
  deployments(input: {projectId: $projectId}, first: 1) {
    edges {
      node {
        id
        status
        updatedAt
        canRedeploy
      }
    }
  }
}`;
    const result = await executeGraphQL(gql, { projectId });

    if (!result.deployments.edges.length) {
        throw new Error(`No deployments found for project ${projectId}`);
    }

    return result.deployments.edges[0].node;
}

async function triggerRedeploy(deploymentId: string): Promise<void> {
    logger.log(`Redeploying deployment: ${deploymentId}`);

    const gql = `
mutation deploymentRedeploy($deploymentId: String!) {
  deploymentRedeploy(id: $deploymentId, usePreviousImageTag: false){
    id
    status
  }
}`;

    const result = await executeGraphQL(gql, { deploymentId });
    logger.log(`Redeployment initiated. Status: ${result.deploymentRedeploy.status}`);
}

async function redeployService(projectId?: string) {
    if(!projectId) {
        throw new Error("Project ID is required for redeployment");
    }
    logger.log(`Redeploying project: ${projectId}`);

    if (!process.env.RAILWAY_TOKEN) {
        throw new Error("RAILWAY_TOKEN environment variable is not set");
    }
    await ensureRailwayInstalled();

    try {
        const { id } = await getLastDeployment(projectId);
        await triggerRedeploy(id);
    } catch (e) {
        logger.error(`Failed to redeploy service for project ${projectId}:`);
        logger.error(e);
    }

    // await asyncExec(`railway redeploy --service "${serviceName}" --yes`);
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

async function executeGraphQL(query: string, variables: any = {}) {
    try {
        const response = await axios({
            url: "https://backboard.railway.com/graphql/v2",
            method: "post",
            headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer ${process.env.RAILWAY_TOKEN}`
            },
            data: {
                query,
                variables
            }
        });

        if (response.data.errors) {
            throw new Error(response.data.errors.map((e: any) => e.message).join(", "));
        }

        return response.data.data;
    } catch (error) {
        logger.error("GraphQL request failed:");
        logger.error(error);
        throw error;
    }
}

export default runExecutor;
