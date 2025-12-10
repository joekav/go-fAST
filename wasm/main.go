package main

import (
	"encoding/json"
	"syscall/js"

	"github.com/t14raptor/go-fast/parser"
	"github.com/t14raptor/go-fast/resolver"
)

func parseJS(this js.Value, args []js.Value) any {
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

	result, err := json.Marshal(program)
	if err != nil {
		return map[string]any{"error": err.Error()}
	}

	return string(result)
}

func main() {
	js.Global().Set("goFastParse", js.FuncOf(parseJS))
	<-make(chan struct{})
}
