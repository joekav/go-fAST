import { readFileSync } from "fs";
import { fileURLToPath } from "url";
import { dirname, join } from "path";
import type { ParseOptions, ParseResult } from "./types.js";

// @ts-ignore - Go WASM runtime
import "../vendor/wasm_exec.js";

declare const Go: any;
declare global {
  var goFastParse: ((source: string, options?: ParseOptions) => string) | undefined;
}

let initialized = false;
let initPromise: Promise<void> | null = null;

/**
 * Initialize the Go WASM module
 */
export async function init(): Promise<void> {
  if (initialized) return;
  if (initPromise) return initPromise;

  initPromise = (async () => {
    const go = new Go();

    // Get path to WASM file relative to this module
    const __filename = fileURLToPath(import.meta.url);
    const __dirname = dirname(__filename);
    const wasmPath = join(__dirname, "go-fast.wasm");

    const wasmBuffer = readFileSync(wasmPath);
    const wasmModule = await WebAssembly.compile(wasmBuffer);
    const instance = await WebAssembly.instantiate(wasmModule, go.importObject);

    go.run(instance);
    initialized = true;
  })();

  return initPromise;
}

/**
 * Reset the WASM module state (useful for testing)
 */
export function teardown(): void {
  initialized = false;
  initPromise = null;
  globalThis.goFastParse = undefined;
}

/**
 * Parse JavaScript source code into an ESTree-compatible AST
 * @param source - The source code to parse
 * @param options - Parser options
 * @returns The parsed AST or an error object
 */
export async function parse(
  source: string,
  options: ParseOptions = {}
): Promise<ParseResult> {
  await init();

  let result: string;
  try {
    result = globalThis.goFastParse!(source, options);
  } catch (e) {
    return { error: `WASM execution failed: ${e instanceof Error ? e.message : String(e)}` };
  }

  if (typeof result !== "string") {
    return { error: `Unexpected result type from WASM: ${typeof result}` };
  }

  try {
    return JSON.parse(result);
  } catch (e) {
    return { error: `Failed to parse WASM output as JSON: ${e instanceof Error ? e.message : String(e)}` };
  }
}

export * from "./types.js";
