package main

import (
	bus_tracker "github.com/ariyn/bus-tracker"
	lox "github.com/ariyn/lox_interpreter"
)

func main() {
	script := `var x = get("https://example.org", "/html/body/div/h1");
print "RESULT " + x;
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

	_, err = interpreter.Interpret(statements)
	if err != nil {
		panic(err)
	}
}
