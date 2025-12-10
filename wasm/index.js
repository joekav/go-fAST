const fs = require("fs");
const path = require("path");

let goFastParse = null;

/**
 * Initialize the Go WASM module
 * @returns {Promise<void>}
 */
async function init() {
  if (goFastParse) return;

  require("./wasm_exec.js");
  const go = new Go();

  const wasmPath = path.join(__dirname, "go-fast.wasm");
  const wasmBuffer = fs.readFileSync(wasmPath);
  const result = await WebAssembly.instantiate(wasmBuffer, go.importObject);

  go.run(result.instance);
  goFastParse = globalThis.goFastParse;
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

module.exports = { init, parse, parseAsync };
