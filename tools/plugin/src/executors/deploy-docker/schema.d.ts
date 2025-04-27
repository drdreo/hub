export interface DeployDockerExecutorSchema {
    projectId?: string;
    dockerFile?: string;
    redeploy?: boolean;
    registry?: string;
    imageTag?: string;
}
