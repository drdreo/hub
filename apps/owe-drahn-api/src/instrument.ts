import * as Sentry from "@sentry/nestjs";
import { nodeProfilingIntegration } from "@sentry/profiling-node";

if (process.env.NODE_ENV === "production") {
    Sentry.init({
        dsn: "https://1f3a7989593230de0b96d41d05b1f5b0@o528779.ingest.us.sentry.io/4508902216892416",
        integrations: [nodeProfilingIntegration()],
        tracesSampleRate: 1.0,
        profilesSampleRate: 0.1
    });

    // Manually call startProfiler and stopProfiler to profile the code in between
    Sentry.profiler.startProfiler();
}
