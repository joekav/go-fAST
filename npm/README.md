# @aspect/go-fast

A fast JavaScript parser written in Go, compiled to WebAssembly. Outputs ESTree-compatible AST.

## Installation

```bash
npm install @aspect/go-fast
```

## Usage

### Node.js

```javascript
import { parseAsync } from "@aspect/go-fast";

const ast = await parseAsync("const x = 1 + 2;");
console.log(ast);
```

### Browser

```javascript
import { init, parse } from "@aspect/go-fast";

// Initialize WASM (fetches from CDN)
await init();

// Parse code
const ast = parse("const x = 1 + 2;");
console.log(ast);
```

### With Scope Resolution

```javascript
import { parseAsync } from "@aspect/go-fast";

const ast = await parseAsync("function foo(x) { return x; }", {
  resolve: true,
});

// Identifiers will have scopeContext field
```

## API

### `init(wasmUrl?: string): Promise<void>`

Initialize the WASM module. In browser environments, optionally provide a custom WASM URL.

### `parse(source: string, options?: ParseOptions): ParseResult`

Synchronously parse source code. Must call `init()` first.

### `parseAsync(source: string, options?: ParseOptions): Promise<ParseResult>`

Initialize (if needed) and parse in one call.

### Options

- `resolve?: boolean` - Enable scope resolution (adds `scopeContext` to identifiers)

## Output Format

Returns ESTree-compatible AST with:
- `type` - Node type (e.g., "Program", "Identifier", "BinaryExpression")
- `start` / `end` - Character offsets
- `scopeContext` - Scope identifier (when `resolve: true`)

## License

MIT
