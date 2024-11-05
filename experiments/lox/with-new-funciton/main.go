package main

import (
	"github.com/ariyn/bus-tracker"
	lox "github.com/ariyn/lox_interpreter"
	"log"
)

func init() {
	lox.NO_RETURN_AT_ROOT = false
}

func main() {
	script := `var x = get("https://example.org", "/html/body/div/h1");
print "crawled " + x;
return {
	title: x
};
`

	scanner := lox.NewScanner(script)
	tokens, _ := scanner.ScanTokens()

	parser := lox.NewParser(tokens)
	statements, _ := parser.Parse()

	env := lox.NewEnvironment(nil)
	env.Define("get", &bus_tracker.GetFunction{})
	interpreter := lox.NewInterpreter(env)

	resolver := lox.NewResolver(interpreter)
	err := resolver.Resolve(statements...)
	if err != nil {
		panic(err)
	}

	v, err := interpreter.Interpret(statements)
	if err != nil {
		panic(err)
	}

	log.Println("RESULT", v)
}
