export interface ReleaseNotificationExecutorSchema {
    /**
     * The version identifier of the release
     */
    version: string;

    /**
     * The project you want to create the release for
     */
    project: string;

    /**
     * The deployment url
     */
    url?: string;
}
