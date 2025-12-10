import { defineConfig } from "tsup";
import { readFileSync, writeFileSync } from "fs";

const pkg = JSON.parse(readFileSync("package.json", "utf-8"));

export default defineConfig([
  // Node.js builds
  {
    entry: { node: "src/node.ts" },
    format: ["esm", "cjs"],
    dts: true,
    clean: false, // Don't clean - WASM is already there
    outDir: "dist",
    external: ["fs", "path", "url"],
    noExternal: ["../vendor/wasm_exec.js"],
    banner: {
      js: "// @ts-nocheck",
    },
  },
  // Browser builds
  {
    entry: { browser: "src/browser.ts" },
    format: ["esm", "cjs"],
    dts: true,
    clean: false,
    outDir: "dist",
    platform: "browser",
    noExternal: ["../vendor/wasm_exec.js"],
    banner: {
      js: "// @ts-nocheck",
    },
  },
]);
