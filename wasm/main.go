package main

import (
	"fmt"
	"syscall/js"

	"github.com/t14raptor/go-fast/parser"
	"github.com/t14raptor/go-fast/resolver"
	"github.com/t14raptor/go-fast/serializer"
)

func parseJS(this js.Value, args []js.Value) (result any) {
	// Recover from panics to prevent WASM from crashing
	defer func() {
		if r := recover(); r != nil {
			result = map[string]any{"error": fmt.Sprintf("internal error: %v", r)}
		}
	}()

	if len(args) < 1 {
		return map[string]any{"error": "No source code provided"}
	}

	source := args[0].String()

	// Check for options object as second argument
	shouldResolve := false
	if len(args) >= 2 && args[1].Type() == js.TypeObject {
		resolveVal := args[1].Get("resolve")
		if resolveVal.Type() == js.TypeBoolean {
			shouldResolve = resolveVal.Bool()
		}
	}

	program, err := parser.ParseFile(source)
	if err != nil {
		return map[string]any{"error": err.Error()}
	}

	if shouldResolve {
		resolver.Resolve(program)
	}

	return serializer.Serialize(program)
}

func main() {
	js.Global().Set("goFastParse", js.FuncOf(parseJS))
	<-make(chan struct{})
}
