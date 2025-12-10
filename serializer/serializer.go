package serializer

import (
	"strconv"
	"sync"
	"unsafe"

	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/token"
)

// Buffer pool to reduce allocations
var serializerPool = sync.Pool{
	New: func() any {
		s := &Serializer{}
		s.out = make([]byte, 0, 8192) // Pre-allocate 8KB
		return s
	},
}

// Serialize converts an AST node to ESTree-compatible JSON.
func Serialize(node ast.VisitableNode) string {
	s := serializerPool.Get().(*Serializer)
	s.out = s.out[:0] // Reset length, keep capacity
	s.V = s
	node.VisitWith(s)
	result := s.String()
	serializerPool.Put(s)
	return result
}

// Serializer implements the ast.Visitor interface to serialize AST to JSON.
type Serializer struct {
	ast.NoopVisitor
	out []byte
}

// writeStr appends a string to the buffer
func (s *Serializer) writeStr(str string) {
	s.out = append(s.out, str...)
}

// writeByte appends a byte to the buffer
func (s *Serializer) writeByte(b byte) {
	s.out = append(s.out, b)
}

// String returns the buffer as a string without copying
func (s *Serializer) String() string {
	return unsafe.String(unsafe.SliceData(s.out), len(s.out))
}

// Helper to serialize a child node
func (s *Serializer) serialize(node ast.VisitableNode) {
	if node == nil {
		s.writeStr("null")
		return
	}
	node.VisitWith(s)
}

// JSON writing helpers
func (s *Serializer) writeString(str string) {
	s.writeByte('"')
	for i := 0; i < len(str); i++ {
		c := str[i]
		switch c {
		case '"':
			s.writeStr(`\"`)
		case '\\':
			s.writeStr(`\\`)
		case '\n':
			s.writeStr(`\n`)
		case '\r':
			s.writeStr(`\r`)
		case '\t':
			s.writeStr(`\t`)
		default:
			if c < 0x20 {
				s.writeStr(`\u00`)
				s.writeByte("0123456789abcdef"[c>>4])
				s.writeByte("0123456789abcdef"[c&0xf])
			} else {
				s.writeByte(c)
			}
		}
	}
	s.writeByte('"')
}

func (s *Serializer) writeNumber(n float64) {
	// Fast path for integers
	if n == float64(int64(n)) && n >= -1e15 && n <= 1e15 {
		s.writeInt64(int64(n))
		return
	}
	s.writeStr(strconv.FormatFloat(n, 'f', -1, 64))
}

// Small int buffer to avoid allocations for common cases
var smallInts = [100]string{
	"0", "1", "2", "3", "4", "5", "6", "7", "8", "9",
	"10", "11", "12", "13", "14", "15", "16", "17", "18", "19",
	"20", "21", "22", "23", "24", "25", "26", "27", "28", "29",
	"30", "31", "32", "33", "34", "35", "36", "37", "38", "39",
	"40", "41", "42", "43", "44", "45", "46", "47", "48", "49",
	"50", "51", "52", "53", "54", "55", "56", "57", "58", "59",
	"60", "61", "62", "63", "64", "65", "66", "67", "68", "69",
	"70", "71", "72", "73", "74", "75", "76", "77", "78", "79",
	"80", "81", "82", "83", "84", "85", "86", "87", "88", "89",
	"90", "91", "92", "93", "94", "95", "96", "97", "98", "99",
}

func (s *Serializer) writeInt(n int) {
	if n >= 0 && n < 100 {
		s.writeStr(smallInts[n])
		return
	}
	s.writeStr(strconv.Itoa(n))
}

func (s *Serializer) writeInt64(n int64) {
	if n >= 0 && n < 100 {
		s.writeStr(smallInts[n])
		return
	}
	s.writeStr(strconv.FormatInt(n, 10))
}

func (s *Serializer) writeBool(b bool) {
	if b {
		s.writeStr("true")
	} else {
		s.writeStr("false")
	}
}

func (s *Serializer) writeNull() {
	s.writeStr("null")
}

// Pre-cached quoted operator strings
var operatorStrings = map[token.Token]string{
	token.Plus:                     `"+"`,
	token.Minus:                    `"-"`,
	token.Multiply:                 `"*"`,
	token.Exponent:                 `"**"`,
	token.Slash:                    `"/"`,
	token.Remainder:                `"%"`,
	token.And:                      `"&"`,
	token.Or:                       `"|"`,
	token.ExclusiveOr:              `"^"`,
	token.ShiftLeft:                `"<<"`,
	token.ShiftRight:               `">>"`,
	token.UnsignedShiftRight:       `">>>"`,
	token.LogicalAnd:               `"&&"`,
	token.LogicalOr:                `"||"`,
	token.Coalesce:                 `"??"`,
	token.Equal:                    `"=="`,
	token.StrictEqual:              `"==="`,
	token.NotEqual:                 `"!="`,
	token.StrictNotEqual:           `"!=="`,
	token.Less:                     `"<"`,
	token.Greater:                  `">"`,
	token.LessOrEqual:              `"<="`,
	token.GreaterOrEqual:           `">="`,
	token.Increment:                `"++"`,
	token.Decrement:                `"--"`,
	token.Not:                      `"!"`,
	token.BitwiseNot:               `"~"`,
	token.Typeof:                   `"typeof"`,
	token.Void:                     `"void"`,
	token.Delete:                   `"delete"`,
	token.In:                       `"in"`,
	token.InstanceOf:               `"instanceof"`,
	token.Assign:                   `"="`,
	token.AddAssign:                `"+="`,
	token.SubtractAssign:           `"-="`,
	token.MultiplyAssign:           `"*="`,
	token.ExponentAssign:           `"**="`,
	token.QuotientAssign:           `"/="`,
	token.RemainderAssign:          `"%="`,
	token.AndAssign:                `"&="`,
	token.OrAssign:                 `"|="`,
	token.ExclusiveOrAssign:        `"^="`,
	token.ShiftLeftAssign:          `"<<="`,
	token.ShiftRightAssign:         `">>="`,
	token.UnsignedShiftRightAssign: `">>>="`,
}

func (s *Serializer) writeOperator(t token.Token) {
	if str, ok := operatorStrings[t]; ok {
		s.writeStr(str)
	} else {
		s.writeString(t.String())
	}
}

// Convert 1-based Go position to 0-based ESTree position
// Returns 0 if position is unset (was 0)
func toESTreePos(pos ast.Idx) int {
	if pos == 0 {
		return 0
	}
	return int(pos) - 1
}

func (s *Serializer) writePosition(node ast.Node) {
	s.writeStr(`"start":`)
	s.writeInt(toESTreePos(node.Idx0()))
	s.writeStr(`,"end":`)
	s.writeInt(toESTreePos(node.Idx1()))
}

func (s *Serializer) writePositionStartOnly(start ast.Idx) {
	s.writeStr(`"start":`)
	s.writeInt(toESTreePos(start))
}

// Program
func (s *Serializer) VisitProgram(n *ast.Program) {
	s.writeStr(`{"type":"Program","body":[`)
	for i, stmt := range n.Body {
		if i > 0 {
			s.writeStr(",")
		}
		s.serialize(stmt.Stmt)
	}
	s.writeStr("]")
	// Program.Idx0()/Idx1() panic on empty body, so only write position if non-empty
	if len(n.Body) > 0 {
		s.writeStr(",")
		s.writePosition(n)
	}
	s.writeStr("}")
}

// Identifiers
func (s *Serializer) VisitIdentifier(n *ast.Identifier) {
	s.writeStr(`{"type":"Identifier","name":`)
	s.writeString(n.Name)
	if n.ScopeContext != 0 {
		s.writeStr(`,"scopeContext":`)
		s.writeInt(int(n.ScopeContext))
	}
	s.writeStr(",")
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitPrivateIdentifier(n *ast.PrivateIdentifier) {
	s.writeStr(`{"type":"PrivateIdentifier","name":`)
	s.writeString(n.Identifier.Name)
	s.writeStr(",")
	s.writePosition(n)
	s.writeStr("}")
}

// Literals
func (s *Serializer) VisitBooleanLiteral(n *ast.BooleanLiteral) {
	s.writeStr(`{"type":"Literal","value":`)
	s.writeBool(n.Value)
	s.writeStr(",")
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitNullLiteral(n *ast.NullLiteral) {
	s.writeStr(`{"type":"Literal","value":null,`)
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitNumberLiteral(n *ast.NumberLiteral) {
	s.writeStr(`{"type":"Literal","value":`)
	s.writeNumber(n.Value)
	if n.Raw != nil {
		s.writeStr(`,"raw":`)
		s.writeString(*n.Raw)
	}
	s.writeStr(",")
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitStringLiteral(n *ast.StringLiteral) {
	s.writeStr(`{"type":"Literal","value":`)
	s.writeString(n.Value)
	if n.Raw != nil {
		s.writeStr(`,"raw":`)
		s.writeString(*n.Raw)
	}
	s.writeStr(",")
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitRegExpLiteral(n *ast.RegExpLiteral) {
	s.writeStr(`{"type":"Literal","regex":{"pattern":`)
	s.writeString(n.Pattern)
	s.writeStr(`,"flags":`)
	s.writeString(n.Flags)
	s.writeStr("},")
	s.writePosition(n)
	s.writeStr("}")
}

// Expressions
func (s *Serializer) VisitBinaryExpression(n *ast.BinaryExpression) {
	s.writeStr(`{"type":"BinaryExpression","operator":`)
	s.writeOperator(n.Operator)
	s.writeStr(`,"left":`)
	s.serialize(n.Left.Expr)
	s.writeStr(`,"right":`)
	s.serialize(n.Right.Expr)
	s.writeStr(",")
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitUnaryExpression(n *ast.UnaryExpression) {
	s.writeStr(`{"type":"UnaryExpression","operator":`)
	s.writeOperator(n.Operator)
	s.writeStr(`,"prefix":true,"argument":`)
	s.serialize(n.Operand.Expr)
	s.writeStr(",")
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitUpdateExpression(n *ast.UpdateExpression) {
	s.writeStr(`{"type":"UpdateExpression","operator":`)
	s.writeOperator(n.Operator)
	s.writeStr(`,"prefix":`)
	s.writeBool(!n.Postfix)
	s.writeStr(`,"argument":`)
	s.serialize(n.Operand.Expr)
	s.writeStr(",")
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitAssignExpression(n *ast.AssignExpression) {
	s.writeStr(`{"type":"AssignmentExpression","operator":`)
	// For compound operators, we need to add "=" suffix
	op := n.Operator.String()
	if n.Operator != token.Assign {
		op += "="
	}
	s.writeString(op)
	s.writeStr(`,"left":`)
	s.serialize(n.Left.Expr)
	s.writeStr(`,"right":`)
	s.serialize(n.Right.Expr)
	s.writeStr(",")
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitConditionalExpression(n *ast.ConditionalExpression) {
	s.writeStr(`{"type":"ConditionalExpression","test":`)
	s.serialize(n.Test.Expr)
	s.writeStr(`,"consequent":`)
	s.serialize(n.Consequent.Expr)
	s.writeStr(`,"alternate":`)
	s.serialize(n.Alternate.Expr)
	s.writeStr(",")
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitCallExpression(n *ast.CallExpression) {
	s.writeStr(`{"type":"CallExpression","callee":`)
	s.serialize(n.Callee.Expr)
	s.writeStr(`,"arguments":[`)
	for i, arg := range n.ArgumentList {
		if i > 0 {
			s.writeStr(",")
		}
		s.serialize(arg.Expr)
	}
	s.writeStr("],")
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitNewExpression(n *ast.NewExpression) {
	s.writeStr(`{"type":"NewExpression","callee":`)
	s.serialize(n.Callee.Expr)
	s.writeStr(`,"arguments":[`)
	for i, arg := range n.ArgumentList {
		if i > 0 {
			s.writeStr(",")
		}
		s.serialize(arg.Expr)
	}
	s.writeStr("],")
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitMemberExpression(n *ast.MemberExpression) {
	s.writeStr(`{"type":"MemberExpression","object":`)
	s.serialize(n.Object.Expr)
	s.writeStr(`,"property":`)
	s.serialize(n.Property)
	s.writeStr(`,"computed":`)
	// Check if computed
	_, isComputed := n.Property.Prop.(*ast.ComputedProperty)
	s.writeBool(isComputed)
	s.writeStr(",")
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitMemberProperty(n *ast.MemberProperty) {
	switch prop := n.Prop.(type) {
	case *ast.Identifier:
		s.serialize(prop)
	case *ast.ComputedProperty:
		s.serialize(prop.Expr.Expr)
	}
}

func (s *Serializer) VisitArrayLiteral(n *ast.ArrayLiteral) {
	s.writeStr(`{"type":"ArrayExpression","elements":[`)
	for i, elem := range n.Value {
		if i > 0 {
			s.writeStr(",")
		}
		if elem.Expr == nil {
			s.writeNull()
		} else {
			s.serialize(elem.Expr)
		}
	}
	s.writeStr("],")
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitObjectLiteral(n *ast.ObjectLiteral) {
	s.writeStr(`{"type":"ObjectExpression","properties":[`)
	for i, prop := range n.Value {
		if i > 0 {
			s.writeStr(",")
		}
		s.serialize(prop.Prop)
	}
	s.writeStr("],")
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitPropertyKeyed(n *ast.PropertyKeyed) {
	s.writeStr(`{"type":"Property","key":`)
	s.serialize(n.Key.Expr)
	s.writeStr(`,"value":`)
	s.serialize(n.Value.Expr)
	s.writeStr(`,"kind":`)
	s.writeString(string(n.Kind))
	s.writeStr(`,"computed":`)
	s.writeBool(n.Computed)
	s.writeStr(`,"method":`)
	s.writeBool(n.Kind == ast.PropertyKindMethod)
	s.writeStr(`,"shorthand":false,`)
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitPropertyShort(n *ast.PropertyShort) {
	s.writeStr(`{"type":"Property","key":`)
	s.serialize(n.Name)
	s.writeStr(`,"value":`)
	if n.Initializer != nil {
		// Shorthand with default: {x = 1}
		s.writeStr(`{"type":"AssignmentPattern","left":`)
		s.serialize(n.Name)
		s.writeStr(`,"right":`)
		s.serialize(n.Initializer.Expr)
		s.writeStr(",")
		s.writePosition(n)
		s.writeStr("}")
	} else {
		s.serialize(n.Name)
	}
	s.writeStr(`,"kind":"init","computed":false,"method":false,"shorthand":true,`)
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitSequenceExpression(n *ast.SequenceExpression) {
	s.writeStr(`{"type":"SequenceExpression","expressions":[`)
	for i, expr := range n.Sequence {
		if i > 0 {
			s.writeStr(",")
		}
		s.serialize(expr.Expr)
	}
	s.writeStr("],")
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitThisExpression(n *ast.ThisExpression) {
	s.writeStr(`{"type":"ThisExpression",`)
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitSuperExpression(n *ast.SuperExpression) {
	s.writeStr(`{"type":"Super",`)
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitYieldExpression(n *ast.YieldExpression) {
	s.writeStr(`{"type":"YieldExpression","argument":`)
	if n.Argument != nil {
		s.serialize(n.Argument.Expr)
	} else {
		s.writeNull()
	}
	s.writeStr(`,"delegate":`)
	s.writeBool(n.Delegate)
	s.writeStr(",")
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitAwaitExpression(n *ast.AwaitExpression) {
	s.writeStr(`{"type":"AwaitExpression","argument":`)
	s.serialize(n.Argument.Expr)
	s.writeStr(",")
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitSpreadElement(n *ast.SpreadElement) {
	s.writeStr(`{"type":"SpreadElement","argument":`)
	s.serialize(n.Expression.Expr)
	s.writeStr(",")
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitTemplateLiteral(n *ast.TemplateLiteral) {
	s.writeStr(`{"type":"TemplateLiteral","quasis":[`)
	for i, elem := range n.Elements {
		if i > 0 {
			s.writeStr(",")
		}
		s.writeStr(`{"type":"TemplateElement","value":{"raw":`)
		s.writeString(elem.Literal)
		s.writeStr(`,"cooked":`)
		s.writeString(elem.Parsed)
		s.writeStr(`},"tail":`)
		s.writeBool(i == len(n.Elements)-1)
		s.writeStr(",")
		s.writePosition(&elem)
		s.writeStr("}")
	}
	s.writeStr(`],"expressions":[`)
	for i, expr := range n.Expressions {
		if i > 0 {
			s.writeStr(",")
		}
		s.serialize(expr.Expr)
	}
	s.writeStr("],")
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitMetaProperty(n *ast.MetaProperty) {
	s.writeStr(`{"type":"MetaProperty","meta":`)
	s.serialize(n.Meta)
	s.writeStr(`,"property":`)
	s.serialize(n.Property)
	s.writeStr(",")
	s.writePosition(n)
	s.writeStr("}")
}

// Functions
func (s *Serializer) VisitFunctionLiteral(n *ast.FunctionLiteral) {
	s.writeStr(`{"type":"FunctionExpression","id":`)
	if n.Name != nil {
		s.serialize(n.Name)
	} else {
		s.writeNull()
	}
	s.writeStr(`,"params":[`)
	for i, param := range n.ParameterList.List {
		if i > 0 {
			s.writeStr(",")
		}
		s.serializeParam(&param)
	}
	if n.ParameterList.Rest != nil {
		if len(n.ParameterList.List) > 0 {
			s.writeStr(",")
		}
		s.writeStr(`{"type":"RestElement","argument":`)
		s.serialize(n.ParameterList.Rest)
		s.writeStr("}")
	}
	s.writeStr(`],"body":`)
	s.serialize(n.Body)
	s.writeStr(`,"generator":`)
	s.writeBool(n.Generator)
	s.writeStr(`,"async":`)
	s.writeBool(n.Async)
	if n.ScopeContext != 0 {
		s.writeStr(`,"scopeContext":`)
		s.writeInt(int(n.ScopeContext))
	}
	s.writeStr(",")
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitArrowFunctionLiteral(n *ast.ArrowFunctionLiteral) {
	s.writeStr(`{"type":"ArrowFunctionExpression","id":null,"params":[`)
	for i, param := range n.ParameterList.List {
		if i > 0 {
			s.writeStr(",")
		}
		s.serializeParam(&param)
	}
	if n.ParameterList.Rest != nil {
		if len(n.ParameterList.List) > 0 {
			s.writeStr(",")
		}
		s.writeStr(`{"type":"RestElement","argument":`)
		s.serialize(n.ParameterList.Rest)
		s.writeStr("}")
	}
	s.writeStr(`],"body":`)
	s.serialize(n.Body)
	s.writeStr(`,"expression":`)
	// expression is true if body is not a BlockStatement
	_, isBlock := n.Body.Body.(*ast.BlockStatement)
	s.writeBool(!isBlock)
	s.writeStr(`,"async":`)
	s.writeBool(n.Async)
	if n.ScopeContext != 0 {
		s.writeStr(`,"scopeContext":`)
		s.writeInt(int(n.ScopeContext))
	}
	s.writeStr(",")
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) serializeParam(param *ast.VariableDeclarator) {
	if param.Initializer != nil {
		s.writeStr(`{"type":"AssignmentPattern","left":`)
		s.serialize(param.Target.Target)
		s.writeStr(`,"right":`)
		s.serialize(param.Initializer.Expr)
		s.writeStr(",")
		s.writePosition(param)
		s.writeStr("}")
	} else {
		s.serialize(param.Target.Target)
	}
}

func (s *Serializer) VisitConciseBody(n *ast.ConciseBody) {
	s.serialize(n.Body)
}

// Statements
func (s *Serializer) VisitBlockStatement(n *ast.BlockStatement) {
	s.writeStr(`{"type":"BlockStatement","body":[`)
	for i, stmt := range n.List {
		if i > 0 {
			s.writeStr(",")
		}
		s.serialize(stmt.Stmt)
	}
	s.writeStr("]")
	if n.ScopeContext != 0 {
		s.writeStr(`,"scopeContext":`)
		s.writeInt(int(n.ScopeContext))
	}
	s.writeStr(",")
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitExpressionStatement(n *ast.ExpressionStatement) {
	s.writeStr(`{"type":"ExpressionStatement","expression":`)
	s.serialize(n.Expression.Expr)
	s.writeStr(",")
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitEmptyStatement(n *ast.EmptyStatement) {
	s.writeStr(`{"type":"EmptyStatement",`)
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitIfStatement(n *ast.IfStatement) {
	s.writeStr(`{"type":"IfStatement","test":`)
	s.serialize(n.Test.Expr)
	s.writeStr(`,"consequent":`)
	s.serialize(n.Consequent.Stmt)
	s.writeStr(`,"alternate":`)
	if n.Alternate != nil {
		s.serialize(n.Alternate.Stmt)
	} else {
		s.writeNull()
	}
	s.writeStr(",")
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitForStatement(n *ast.ForStatement) {
	s.writeStr(`{"type":"ForStatement","init":`)
	if n.Initializer != nil {
		s.serialize(n.Initializer)
	} else {
		s.writeNull()
	}
	s.writeStr(`,"test":`)
	if n.Test.Expr != nil {
		s.serialize(n.Test.Expr)
	} else {
		s.writeNull()
	}
	s.writeStr(`,"update":`)
	if n.Update.Expr != nil {
		s.serialize(n.Update.Expr)
	} else {
		s.writeNull()
	}
	s.writeStr(`,"body":`)
	s.serialize(n.Body.Stmt)
	s.writeStr(",")
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitForLoopInitializer(n *ast.ForLoopInitializer) {
	switch init := n.Initializer.(type) {
	case *ast.Expression:
		s.serialize(init.Expr)
	case *ast.VariableDeclaration:
		s.serialize(init)
	}
}

func (s *Serializer) VisitForInStatement(n *ast.ForInStatement) {
	s.writeStr(`{"type":"ForInStatement","left":`)
	s.serialize(n.Into)
	s.writeStr(`,"right":`)
	s.serialize(n.Source.Expr)
	s.writeStr(`,"body":`)
	s.serialize(n.Body.Stmt)
	s.writeStr(",")
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitForOfStatement(n *ast.ForOfStatement) {
	s.writeStr(`{"type":"ForOfStatement","left":`)
	s.serialize(n.Into)
	s.writeStr(`,"right":`)
	s.serialize(n.Source.Expr)
	s.writeStr(`,"body":`)
	s.serialize(n.Body.Stmt)
	s.writeStr(`,"await":false,`)
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitForInto(n *ast.ForInto) {
	switch into := n.Into.(type) {
	case *ast.VariableDeclaration:
		s.serialize(into)
	case *ast.Expression:
		s.serialize(into.Expr)
	}
}

func (s *Serializer) VisitWhileStatement(n *ast.WhileStatement) {
	s.writeStr(`{"type":"WhileStatement","test":`)
	s.serialize(n.Test.Expr)
	s.writeStr(`,"body":`)
	s.serialize(n.Body.Stmt)
	s.writeStr(",")
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitDoWhileStatement(n *ast.DoWhileStatement) {
	s.writeStr(`{"type":"DoWhileStatement","body":`)
	s.serialize(n.Body.Stmt)
	s.writeStr(`,"test":`)
	s.serialize(n.Test.Expr)
	s.writeStr(",")
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitSwitchStatement(n *ast.SwitchStatement) {
	s.writeStr(`{"type":"SwitchStatement","discriminant":`)
	s.serialize(n.Discriminant.Expr)
	s.writeStr(`,"cases":[`)
	for i := range n.Body {
		if i > 0 {
			s.writeStr(",")
		}
		s.serialize(&n.Body[i])
	}
	s.writeStr("],")
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitCaseStatement(n *ast.CaseStatement) {
	s.writeStr(`{"type":"SwitchCase","test":`)
	if n.Test != nil {
		s.serialize(n.Test.Expr)
	} else {
		s.writeNull()
	}
	s.writeStr(`,"consequent":[`)
	for i, stmt := range n.Consequent {
		if i > 0 {
			s.writeStr(",")
		}
		s.serialize(stmt.Stmt)
	}
	s.writeStr("],")
	// CaseStatement.Idx1() panics on empty Consequent, so use start only for fallthrough cases
	if len(n.Consequent) > 0 {
		s.writePosition(n)
	} else {
		s.writePositionStartOnly(n.Case)
	}
	s.writeStr("}")
}

func (s *Serializer) VisitTryStatement(n *ast.TryStatement) {
	s.writeStr(`{"type":"TryStatement","block":`)
	s.serialize(n.Body)
	s.writeStr(`,"handler":`)
	if n.Catch != nil {
		s.serialize(n.Catch)
	} else {
		s.writeNull()
	}
	s.writeStr(`,"finalizer":`)
	if n.Finally != nil {
		s.serialize(n.Finally)
	} else {
		s.writeNull()
	}
	s.writeStr(",")
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitCatchStatement(n *ast.CatchStatement) {
	s.writeStr(`{"type":"CatchClause","param":`)
	if n.Parameter != nil && n.Parameter.Target != nil {
		s.serialize(n.Parameter.Target)
	} else {
		s.writeNull()
	}
	s.writeStr(`,"body":`)
	s.serialize(n.Body)
	s.writeStr(",")
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitThrowStatement(n *ast.ThrowStatement) {
	s.writeStr(`{"type":"ThrowStatement","argument":`)
	s.serialize(n.Argument.Expr)
	s.writeStr(",")
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitReturnStatement(n *ast.ReturnStatement) {
	s.writeStr(`{"type":"ReturnStatement","argument":`)
	if n.Argument != nil {
		s.serialize(n.Argument.Expr)
	} else {
		s.writeNull()
	}
	s.writeStr(",")
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitBreakStatement(n *ast.BreakStatement) {
	s.writeStr(`{"type":"BreakStatement","label":`)
	if n.Label != nil {
		s.serialize(n.Label)
	} else {
		s.writeNull()
	}
	s.writeStr(",")
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitContinueStatement(n *ast.ContinueStatement) {
	s.writeStr(`{"type":"ContinueStatement","label":`)
	if n.Label != nil {
		s.serialize(n.Label)
	} else {
		s.writeNull()
	}
	s.writeStr(",")
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitLabelledStatement(n *ast.LabelledStatement) {
	s.writeStr(`{"type":"LabeledStatement","label":`)
	s.serialize(n.Label)
	s.writeStr(`,"body":`)
	s.serialize(n.Statement.Stmt)
	s.writeStr(",")
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitWithStatement(n *ast.WithStatement) {
	s.writeStr(`{"type":"WithStatement","object":`)
	s.serialize(n.Object.Expr)
	s.writeStr(`,"body":`)
	s.serialize(n.Body.Stmt)
	s.writeStr(",")
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitDebuggerStatement(n *ast.DebuggerStatement) {
	s.writeStr(`{"type":"DebuggerStatement",`)
	s.writePosition(n)
	s.writeStr("}")
}

// Declarations
func (s *Serializer) VisitVariableDeclaration(n *ast.VariableDeclaration) {
	s.writeStr(`{"type":"VariableDeclaration","declarations":[`)
	for i := range n.List {
		if i > 0 {
			s.writeStr(",")
		}
		s.serialize(&n.List[i])
	}
	s.writeStr(`],"kind":`)
	s.writeString(n.Token.String())
	s.writeStr(",")
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitVariableDeclarator(n *ast.VariableDeclarator) {
	s.writeStr(`{"type":"VariableDeclarator","id":`)
	s.serialize(n.Target.Target)
	s.writeStr(`,"init":`)
	if n.Initializer != nil {
		s.serialize(n.Initializer.Expr)
	} else {
		s.writeNull()
	}
	s.writeStr(",")
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitFunctionDeclaration(n *ast.FunctionDeclaration) {
	s.writeStr(`{"type":"FunctionDeclaration","id":`)
	if n.Function.Name != nil {
		s.serialize(n.Function.Name)
	} else {
		s.writeNull()
	}
	s.writeStr(`,"params":[`)
	for i, param := range n.Function.ParameterList.List {
		if i > 0 {
			s.writeStr(",")
		}
		s.serializeParam(&param)
	}
	if n.Function.ParameterList.Rest != nil {
		if len(n.Function.ParameterList.List) > 0 {
			s.writeStr(",")
		}
		s.writeStr(`{"type":"RestElement","argument":`)
		s.serialize(n.Function.ParameterList.Rest)
		s.writeStr("}")
	}
	s.writeStr(`],"body":`)
	s.serialize(n.Function.Body)
	s.writeStr(`,"generator":`)
	s.writeBool(n.Function.Generator)
	s.writeStr(`,"async":`)
	s.writeBool(n.Function.Async)
	if n.Function.ScopeContext != 0 {
		s.writeStr(`,"scopeContext":`)
		s.writeInt(int(n.Function.ScopeContext))
	}
	s.writeStr(",")
	s.writePosition(n)
	s.writeStr("}")
}

// Patterns
func (s *Serializer) VisitArrayPattern(n *ast.ArrayPattern) {
	s.writeStr(`{"type":"ArrayPattern","elements":[`)
	for i, elem := range n.Elements {
		if i > 0 {
			s.writeStr(",")
		}
		if elem.Expr == nil {
			s.writeNull()
		} else {
			s.serialize(elem.Expr)
		}
	}
	if n.Rest != nil {
		if len(n.Elements) > 0 {
			s.writeStr(",")
		}
		s.writeStr(`{"type":"RestElement","argument":`)
		s.serialize(n.Rest.Expr)
		s.writeStr("}")
	}
	s.writeStr("],")
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitObjectPattern(n *ast.ObjectPattern) {
	s.writeStr(`{"type":"ObjectPattern","properties":[`)
	for i, prop := range n.Properties {
		if i > 0 {
			s.writeStr(",")
		}
		s.serialize(prop.Prop)
	}
	if n.Rest != nil {
		if len(n.Properties) > 0 {
			s.writeStr(",")
		}
		s.writeStr(`{"type":"RestElement","argument":`)
		s.serialize(n.Rest)
		s.writeStr("}")
	}
	s.writeStr("],")
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitBindingTarget(n *ast.BindingTarget) {
	s.serialize(n.Target)
}

// Classes
func (s *Serializer) VisitClassLiteral(n *ast.ClassLiteral) {
	s.writeStr(`{"type":"ClassExpression","id":`)
	if n.Name != nil {
		s.serialize(n.Name)
	} else {
		s.writeNull()
	}
	s.writeStr(`,"superClass":`)
	if n.SuperClass != nil {
		s.serialize(n.SuperClass.Expr)
	} else {
		s.writeNull()
	}
	s.writeStr(`,"body":{"type":"ClassBody","body":[`)
	for i, elem := range n.Body {
		if i > 0 {
			s.writeStr(",")
		}
		s.serialize(elem.Element)
	}
	s.writeStr("]},")
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitClassDeclaration(n *ast.ClassDeclaration) {
	s.writeStr(`{"type":"ClassDeclaration","id":`)
	if n.Class.Name != nil {
		s.serialize(n.Class.Name)
	} else {
		s.writeNull()
	}
	s.writeStr(`,"superClass":`)
	if n.Class.SuperClass != nil {
		s.serialize(n.Class.SuperClass.Expr)
	} else {
		s.writeNull()
	}
	s.writeStr(`,"body":{"type":"ClassBody","body":[`)
	for i, elem := range n.Class.Body {
		if i > 0 {
			s.writeStr(",")
		}
		s.serialize(elem.Element)
	}
	s.writeStr("]},")
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitMethodDefinition(n *ast.MethodDefinition) {
	s.writeStr(`{"type":"MethodDefinition","key":`)
	s.serialize(n.Key.Expr)
	s.writeStr(`,"value":`)
	s.serialize(n.Body)
	s.writeStr(`,"kind":`)
	kind := string(n.Kind)
	if kind == "method" {
		kind = "method"
	} else if kind == "" {
		kind = "method"
	}
	s.writeString(kind)
	s.writeStr(`,"computed":`)
	s.writeBool(n.Computed)
	s.writeStr(`,"static":`)
	s.writeBool(n.Static)
	s.writeStr(",")
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitFieldDefinition(n *ast.FieldDefinition) {
	s.writeStr(`{"type":"PropertyDefinition","key":`)
	s.serialize(n.Key.Expr)
	s.writeStr(`,"value":`)
	if n.Initializer != nil {
		s.serialize(n.Initializer.Expr)
	} else {
		s.writeNull()
	}
	s.writeStr(`,"computed":`)
	s.writeBool(n.Computed)
	s.writeStr(`,"static":`)
	s.writeBool(n.Static)
	s.writeStr(",")
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitClassStaticBlock(n *ast.ClassStaticBlock) {
	s.writeStr(`{"type":"StaticBlock","body":[`)
	for i, stmt := range n.Block.List {
		if i > 0 {
			s.writeStr(",")
		}
		s.serialize(stmt.Stmt)
	}
	s.writeStr("],")
	s.writePosition(n)
	s.writeStr("}")
}

// Optional chaining
func (s *Serializer) VisitOptionalChain(n *ast.OptionalChain) {
	s.writeStr(`{"type":"ChainExpression","expression":`)
	s.serialize(n.Base.Expr)
	s.writeStr(",")
	s.writePosition(n)
	s.writeStr("}")
}

func (s *Serializer) VisitOptional(n *ast.Optional) {
	s.serialize(n.Expr.Expr)
}

func (s *Serializer) VisitPrivateDotExpression(n *ast.PrivateDotExpression) {
	s.writeStr(`{"type":"MemberExpression","object":`)
	s.serialize(n.Left.Expr)
	s.writeStr(`,"property":`)
	s.serialize(n.Identifier)
	s.writeStr(`,"computed":false,`)
	s.writePosition(n)
	s.writeStr("}")
}
