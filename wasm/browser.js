import "./wasm_exec.js";

let goFastParse = null;
let initPromise = null;

/**
 * Initialize the Go WASM module for browser environments
 * @param {string} [wasmUrl] - URL to the WASM file (defaults to fetching from package)
 * @returns {Promise<void>}
 */
async function init(wasmUrl) {
    if (goFastParse) return;
    if (initPromise) return initPromise;

    initPromise = (async () => {
        const go = new Go();

        const wasmPath = wasmUrl || new URL("go-fast.wasm", import.meta.url).href;
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
