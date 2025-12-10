import "./wasm_exec.js";
import { name as PKG_NAME, version as PKG_VERSION } from "./package.json";

let goFastParse = null;
let initPromise = null;

/**
 * Initialize the Go WASM module for browser environments
 * @param {string} [customWasmUrl] - URL to the WASM file (defaults to unpkg CDN)
 * @returns {Promise<void>}
 */
async function init(customWasmUrl) {
    if (goFastParse) return;
    if (initPromise) return initPromise;

    initPromise = (async () => {
        const go = new Go();

        const wasmPath = customWasmUrl || `https://unpkg.com/${PKG_NAME}@${PKG_VERSION}/go-fast.wasm`;
        const result = await WebAssembly.instantiateStreaming(fetch(wasmPath), go.importObject);

        go.run(result.instance);
        goFastParse = globalThis.goFastParse;
    })();

    return initPromise;
}

/**
 * Parse JavaScript source code into an AST
 * @param {string} source - The JavaScript source code to parse
 * @param {Object} [options] - Parser options
 * @param {boolean} [options.resolve=false] - Whether to resolve scopes
 * @returns {Object} The parsed AST or an error object
 */
function parse(source, options = {}) {
    if (!goFastParse) {
        throw new Error("WASM not initialized. Call init() first.");
    }

    const result = goFastParse(source, options);
    return JSON.parse(result);
}

/**
 * Initialize and parse in one call
 * @param {string} source - The JavaScript source code to parse
 * @param {Object} [options] - Parser options
 * @param {boolean} [options.resolve=false] - Whether to resolve scopes
 * @returns {Promise<Object>} The parsed AST or an error object
 */
async function parseAsync(source, options = {}) {
    await init();
    return parse(source, options);
}

export { init, parse, parseAsync };
