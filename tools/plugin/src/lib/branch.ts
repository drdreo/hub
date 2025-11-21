/**
 * Utilities for determining branch/event context inside CI runners (GitHub Actions + Nx)
 */

/** Derive current branch name from common CI environment variables. */
export function getCurrentBranch(): string {
    const raw =
        process.env.NX_BRANCH ||
        process.env.GITHUB_HEAD_REF ||
        process.env.GITHUB_REF_NAME ||
        process.env.GITHUB_REF ||
        "";
    return raw.replace(/^refs\/heads\//, "").trim();
}

/** Whether the current branch is a production branch (main/master). */
export function isMainBranch(branch = getCurrentBranch()): boolean {
    return ["main", "master"].includes(branch);
}

/** Whether the current GitHub event is a pull request. */
export function isPullRequest(): boolean {
    return process.env.GITHUB_EVENT_NAME === "pull_request";
}

/** Determine if we should treat this execution as a preview (e.g. PR builds). */
export function shouldPreviewDeploy(branch = getCurrentBranch()): boolean {
    return isPullRequest() && !isMainBranch(branch);
}

export function computeEffectiveDryRun(
    explicitDryRun: boolean | undefined,
    branch = getCurrentBranch()
): boolean {
    if (explicitDryRun) {
        return true;
    }
    if (shouldPreviewDeploy(branch)) {
        return false;
    }
    // safeguard: never push from non-main
    if (!isMainBranch(branch)) {
        return true;
    }
    return false;
}
