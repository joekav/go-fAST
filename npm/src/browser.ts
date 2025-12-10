import type { ParseOptions, ParseResult } from "./types.js";

// @ts-ignore - Go WASM runtime
import "../vendor/wasm_exec.js";

declare const Go: any;
declare global {
    var goFastParse: ((source: string, options?: ParseOptions) => string) | undefined;
}

let initialized = false;
let initPromise: Promise<void> | null = null;

export interface InitOptions {
    /** URL to the go-fast.wasm file. Required in browser environments. */
    wasmURL: string;
}

/**
 * Initialize the Go WASM module for browser environments
 * @param options - Configuration options including wasmURL
 */
export async function init(options: InitOptions): Promise<void> {
    if (initialized) return;
    if (initPromise) return initPromise;

    const { wasmURL } = options;
    if (!wasmURL) {
        throw new Error('Must provide the "wasmURL" option');
    }

    initPromise = (async () => {
        const go = new Go();

        let result: WebAssembly.WebAssemblyInstantiatedSource;
        if (typeof WebAssembly.instantiateStreaming === "function") {
            result = await WebAssembly.instantiateStreaming(fetch(wasmURL), go.importObject);
        } else {
            // Fallback for older browsers
            const response = await fetch(wasmURL);
            const bytes = await response.arrayBuffer();
            result = await WebAssembly.instantiate(bytes, go.importObject);
        }

        go.run(result.instance);
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
export function parse(source: string, options: ParseOptions = {}): ParseResult {
    if (!initialized || !globalThis.goFastParse) {
        throw new Error("WASM not initialized. Call init() first.");
    }

    let result: string;
    try {
        result = globalThis.goFastParse(source, options);
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
