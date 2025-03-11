const { NxAppWebpackPlugin } = require("@nx/webpack/app-plugin");
const { join } = require("path");

module.exports = {
    output: {
        path: join(__dirname, "dist"),
        filename: "[name].cjs" // added to force cjs output
    },
    resolve: {
        extensions: [".ts", ".js", ".json"],
        extensionAlias: {
            ".js": [".ts", ".js"]
        }
    },
    plugins: [
        new NxAppWebpackPlugin({
            target: "node",
            compiler: "tsc",
            main: "./src/main.ts",
            tsConfig: "./tsconfig.app.json",
            assets: ["./src/assets"],
            optimization: false,
            outputHashing: "none",
            generatePackageJson: false // true
        })
    ]
};
