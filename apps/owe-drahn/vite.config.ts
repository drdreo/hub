/// <reference types='vitest' />
import { sentryVitePlugin } from "@sentry/vite-plugin";
import react from "@vitejs/plugin-react-swc";
import { defineConfig } from "vite";
import { ViteImageOptimizer } from "vite-plugin-image-optimizer";

export default defineConfig({
    root: __dirname,
    cacheDir: "../../node_modules/.vite/apps/owe-drahn",
    server: {
        port: 4200,
        host: "localhost"
    },
    preview: {
        port: 4300,
        host: "localhost"
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
            disable: !!process.env.VITEST
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
});
