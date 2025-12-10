export interface ParseOptions {
  /** Whether to resolve variable scopes */
  resolve?: boolean;
}

export interface Position {
  start: number;
  end: number;
}

export interface BaseNode extends Position {
  type: string;
}

export interface Identifier extends BaseNode {
  type: "Identifier";
  name: string;
  scopeContext?: number;
}

export interface Literal extends BaseNode {
  type: "Literal";
  value: string | number | boolean | null;
  raw?: string;
  regex?: { pattern: string; flags: string };
}

export interface Program extends BaseNode {
  type: "Program";
  body: Statement[];
}

// ESTree compatible types
export type Statement = BaseNode;
export type Expression = BaseNode;
export type Node = Program | Statement | Expression | Identifier | Literal;

export interface ParseError {
  error: string;
}

export type ParseResult = Program | ParseError;

export function isParseError(result: ParseResult): result is ParseError {
  return "error" in result;
}
