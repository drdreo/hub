export interface DeployDockerExecutorSchema {
    serviceName: string;
    dockerFile?: string;
    redeploy?: boolean;
    registry?: string;
    imageTag?: string;
}
