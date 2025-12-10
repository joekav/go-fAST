export interface ParseOptions {
  /** Whether to resolve variable scopes */
  resolve?: boolean;
}

export interface Position {
  Line: number;
  Column: number;
}

export interface SourceLocation {
  Start: Position;
  End: Position;
}

export interface Node {
  Idx0: number;
  Idx1: number;
}

export interface Identifier extends Node {
  Name: string;
}

export interface Program extends Node {
  Body: Statement[];
}

export type Statement = Node;
export type Expression = Node;

export interface ParseError {
  error: string;
}

/**
 * Initialize the WASM module. Must be called before parse().
 */
export function init(): Promise<void>;

/**
 * Parse JavaScript source code into an AST.
 * Requires init() to be called first.
 */
export function parse(source: string, options?: ParseOptions): Program | ParseError;

/**
 * Initialize (if needed) and parse JavaScript source code.
 * Convenience method that combines init() and parse().
 */
export function parseAsync(source: string, options?: ParseOptions): Promise<Program | ParseError>;
