package main

import (
	lox "github.com/ariyn/lox_interpreter"
	"log"
)

func main() {
	script := `
var a = 1;
var b = 2;
print a + b;
`
	scanner := lox.NewScanner(script)
	tokens, _ := scanner.ScanTokens()

	parser := lox.NewParser(tokens)
	statements, _ := parser.Parse()

	interpreter := lox.NewInterpreter()

	resolver := lox.NewResolver(interpreter)
	resolver.Resolve(statements...)

	v, _ := interpreter.Interpret(statements)

	log.Println(v)
}
