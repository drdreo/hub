/// <reference types='vitest' />
import { sentryVitePlugin } from "@sentry/vite-plugin";
import react from "@vitejs/plugin-react-swc";
import { defineConfig } from "vitest/config";
import { ViteImageOptimizer } from "vite-plugin-image-optimizer";

export default defineConfig(({ mode }) => ({
    root: __dirname,
    cacheDir: "../../node_modules/.vite/apps/owe-drahn",
    mode,
    envDir: __dirname,
    server: {
        port: 4200,
        host: true
    },
    plugins: [
        react(),
        ViteImageOptimizer({
            png: { quality: 80 },
            jpeg: { quality: 75 },
            webp: { quality: 80 },
            avif: { quality: 70 }
        }),
        sentryVitePlugin({
            org: "drdreo",
            project: "owe-drahn",
            disable: process.env.NODE_ENV !== "production"
        })
    ],
    // Uncomment this if you are using workers.
    // worker: {
    //  plugins: [ nxViteTsPaths() ],
    // },
    build: {
        outDir: "./dist",
        emptyOutDir: true,
        reportCompressedSize: true,
        sourcemap: true,
        commonjsOptions: {
            transformMixedEsModules: true
        }
    },
    test: {
        watch: false,
        globals: true,
        environment: "jsdom",
        passWithNoTests: true,
        include: ["src/**/*.{test,spec}.{js,mjs,cjs,ts,mts,cts,jsx,tsx}"],
        reporters: ["default"],
        coverage: {
            reportsDirectory: "./test-output/vitest/coverage",
            provider: "v8"
        }
    }
}));
