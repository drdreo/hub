import { ExecutorContext, PromiseExecutor, runExecutor as nxRunExecutor } from "@nx/devkit";
import { execSync } from "child_process";
import type { CloudflareExecutorSchema } from "./schema.d.ts";
import {
    getCurrentBranch,
    isMainBranch,
    shouldPreviewDeploy,
    computeEffectiveDryRun
} from "../../lib/branch.ts";

const runExecutor: PromiseExecutor<CloudflareExecutorSchema> = async (
    options,
    context: ExecutorContext
) => {
    const { buildTarget, config, dryRun: explicitDryRun, env } = options; // removed unused dist

    const branch = getCurrentBranch();
    const preview = shouldPreviewDeploy(branch);
    const dryRun = computeEffectiveDryRun(explicitDryRun, branch);

    if (preview) {
        console.log(`\n🔍 Preview deployment for PR on branch '${branch}'.`);
    } else if (!isMainBranch(branch)) {
        console.log(`\n⚠ Cloudflare deploy on non-main branch '${branch}'. Will run in dry-run mode.`);
    }

    try {
        // Build the project if buildTarget is specified
        if (buildTarget) {
            console.log(`Building target: ${buildTarget}`);

            const [project, target, configuration] = buildTarget.split(":");

            for await (const output of await nxRunExecutor(
                { project, target, configuration },
                {},
                context
            )) {
                if (!output.success) {
                    console.error(`Build failed for ${buildTarget}`);
                    return { success: false };
                }
            }

            console.log(`✓ Build completed successfully`);
        }

        // Prepare wrangler command
        const wranglerArgs = ["deploy"];

        if (config) {
            wranglerArgs.push("--config", config);
        }

        // For PR preview we ask wrangler for a preview via --dry-run (it will output URLs)
        if (dryRun) {
            wranglerArgs.push("--dry-run");
        }

        if (env) {
            wranglerArgs.push("--env", env);
        }

        const command = `pnpm dlx wrangler ${wranglerArgs.join(" ")}`;

        console.log(`\nDeploying to Cloudflare...`);
        console.log(`Command: ${command}\n`);

        execSync(command, {
            stdio: "inherit",
            cwd: context.root
        });

        if (dryRun) {
            console.log("\n✓ Dry-run / preview completed (no production resources modified).");
            return { success: true };
        }

        console.log("\n✓ Successfully deployed to Cloudflare!");

        return { success: true };
    } catch (error) {
        console.error("Deployment failed:", error);
        return { success: false };
    }
};

export default runExecutor;
