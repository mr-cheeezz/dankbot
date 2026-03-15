import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import { createHtmlPlugin } from "vite-plugin-html";
export default defineConfig({
    plugins: [
        react(),
        createHtmlPlugin({
            minify: {
                collapseWhitespace: true,
                keepClosingSlash: true,
                removeComments: true,
                removeRedundantAttributes: true,
                removeScriptTypeAttributes: true,
                removeStyleLinkTypeAttributes: true,
                useShortDoctype: true,
                minifyCSS: true,
                minifyJS: true,
            },
        }),
    ],
    server: {
        port: 5173,
    },
    build: {
        target: "es2019",
        minify: "esbuild",
        cssMinify: true,
        sourcemap: false,
        modulePreload: {
            polyfill: true,
        },
        rollupOptions: {
            output: {
                manualChunks: function (id) {
                    if (id.indexOf("node_modules/react-router") !== -1) {
                        return "router";
                    }
                    if (id.indexOf("node_modules/react") !== -1 ||
                        id.indexOf("node_modules/react-dom") !== -1 ||
                        id.indexOf("node_modules/@mui") !== -1 ||
                        id.indexOf("node_modules/@emotion") !== -1) {
                        return "vendor";
                    }
                    return undefined;
                },
            },
        },
    },
    esbuild: {
        drop: ["console", "debugger"],
        legalComments: "none",
    },
});
