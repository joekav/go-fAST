package serializer

import (
	"encoding/json"
	"testing"

	"github.com/t14raptor/go-fast/parser"
)

// Comprehensive JS code covering all supported features
const comprehensiveJS = `
// Variables
var a = 1;
let b = 2;
const c = 3;

// Literals
const num = 42;
const float = 3.14;
const str = "hello";
const str2 = 'world';
const bool1 = true;
const bool2 = false;
const nil = null;
const regex = /pattern/gi;

// Template literals
const template = ` + "`hello ${name} world`" + `;

// Arrays
const arr = [1, 2, 3];
const sparse = [1, , 3];
const spread = [...arr, 4, 5];

// Objects
const obj = {
    a: 1,
    b: 2,
    "quoted": 3,
    [computed]: 4,
    get x() { return 1; },
    set x(v) { },
    method() { return 1; }
};

// Destructuring
const { x: renamed, z: z2 } = obj;
const [first, second, ...rest] = arr;

// Functions
function named(a, b) {
    return a + b;
}

const arrow = (x) => x * 2;
const arrowBlock = (x) => { return x * 2; };
const arrowImplicit = x => x * 2;

async function asyncFn() {
    await promise;
    return 1;
}

function* generator() {
    yield 1;
    yield* other;
}

function withDefaults(a = 1, b = 2) {}
function withRest(a, ...args) {}

// Classes
class Animal {
    constructor(name) {
        this.name = name;
    }

    speak() {
        return this.name;
    }

    static create() {
        return new Animal("default");
    }
}

class Dog extends Animal {
    constructor(name) {
        super(name);
    }

    speak() {
        return super.speak() + " barks";
    }
}

// Expressions
const binary = 1 + 2 - 3 * 4 / 5 % 6;
const comparison = a < b && b > c || a <= b && b >= c;
const equality = a == b && a === b && a != b && a !== b;
const bitwise = a & b | c ^ d << 1 >> 2 >>> 3;
const logical = (a && b) || c;
const unary = !a && -b && +c && ~d && typeof e && void f && delete g;
const update = ++a + b++ + --c + d--;
const ternary = a ? b : c;
a = b;
a += c;
a -= d;
a *= e;
a /= f;
a %= g;
const sequence = (a, b, c);
const call = fn(1, 2, 3);
const callSpread = fn(...args);
const newExpr = new Class(1, 2);
const member = obj.prop;
const computed2 = obj[prop];
const chain = obj?.prop?.method?.();
const newTarget = function() { return new.target; };

// Statements
if (a) {
    b;
} else if (c) {
    d;
} else {
    e;
}

for (let i = 0; i < 10; i++) {
    continue;
}

for (const key in obj) {
    break;
}

for (const item of arr) {
    item;
}

while (condition) {
    something;
}

do {
    something;
} while (condition);

switch (value) {
    case 1:
        one;
        break;
    case 2:
    case 3:
        twoOrThree;
        break;
    default:
        other;
}

try {
    risky();
} catch (e) {
    handle(e);
} finally {
    cleanup();
}

try {
    risky();
} catch {
    handleNoParam();
}

label: for (;;) {
    break label;
    continue label;
}

with (obj) {
    prop;
}

throw new Error("test");

debugger;

// This and super
class Test {
    method() {
        this.prop;
        super.method();
    }
}

// Meta properties
function checkNew() {
    if (new.target) {
        return true;
    }
}
`

func TestComprehensiveJS(t *testing.T) {
	program, err := parser.ParseFile(comprehensiveJS)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	result := Serialize(program)

	// Verify it's valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("Invalid JSON output: %v\nOutput: %s", err, result)
	}

	// Check it has the expected structure
	if parsed["type"] != "Program" {
		t.Errorf("Expected type 'Program', got %v", parsed["type"])
	}

	body, ok := parsed["body"].([]interface{})
	if !ok {
		t.Fatalf("Expected body to be array")
	}

	if len(body) == 0 {
		t.Error("Expected non-empty body")
	}

	t.Logf("Successfully serialized %d statements", len(body))
}

func BenchmarkSerializerCustom(b *testing.B) {
	program, err := parser.ParseFile(comprehensiveJS)
	if err != nil {
		b.Fatalf("Parse error: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Serialize(program)
	}
}

func BenchmarkSerializerJSON(b *testing.B) {
	program, err := parser.ParseFile(comprehensiveJS)
	if err != nil {
		b.Fatalf("Parse error: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(program)
	}
}

// Also benchmark a simpler case
const simpleJS = `const x = 1 + 2;`

func BenchmarkSerializerCustomSimple(b *testing.B) {
	program, err := parser.ParseFile(simpleJS)
	if err != nil {
		b.Fatalf("Parse error: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Serialize(program)
	}
}

func BenchmarkSerializerJSONSimple(b *testing.B) {
	program, err := parser.ParseFile(simpleJS)
	if err != nil {
		b.Fatalf("Parse error: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(program)
	}
}
