import { ExecutorContext, logger } from "@nx/devkit";
import axios from "axios";
import { ReleaseNotificationExecutorSchema } from "./schema.js";

/**
 * Executor to notify Sentry about a new release deployment
 * Based on Sentry API: https://docs.sentry.io/api/releases/create-a-new-release-for-an-organization/
 */
const runExecutor = async (options: ReleaseNotificationExecutorSchema, context: ExecutorContext) => {
    try {
        const { project, url } = options;

        let version: string | undefined = options.version;
        if (!version) {
            version = process.env.GIT_HASH;
        }

        if (!version) {
            throw new Error("Version not provided and GIT_HASH environment variable not found");
        }
        logger.info(`Preparing release notification for '${project}' - version: ${version}`);

        const payload: Record<string, any> = {
            version,
            projects: [project]
        };

        if (url) payload.url = url;

        const authToken =
            process.env.SENTRY_AUTH_TOKEN ||
            "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";
        if (!authToken) {
            throw new Error("Sentry credentials not found");
        }

        logger.info(`Notifying Sentry about release deployment...`);

        const response = await axios({
            method: "POST",
            url: `https://sentry.io/api/0/organizations/drdreo/releases/`,
            headers: {
                Authorization: `Bearer ${authToken}`,
                "Content-Type": "application/json"
            },
            data: payload
        });

        if (response.status >= 200 && response.status < 300) {
            logger.info(`✅ Successfully created Sentry release: ${version}`);
            return {
                success: true,
                message: `Release notification successful for version: ${version}`
            };
        } else {
            throw new Error(`Sentry API returned unexpected status: ${response.status}`);
        }
    } catch (error) {
        if (axios.isAxiosError(error)) {
            const statusCode = error.response?.status;
            const responseData = error.response?.data;
            logger.error(
                `❌ Release Notification failed with status ${statusCode}: ${JSON.stringify(
                    responseData
                )}`
            );
        } else {
            logger.error(`❌ Release Notification failed:` + error);
        }

        return {
            success: false,
            message: error instanceof Error ? error.message : "Unknown error"
        };
    }
};

export default runExecutor;
