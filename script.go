package bus_tracker

import (
	"github.com/ariyn/bus-tracker/functions"
	lox "github.com/ariyn/lox_interpreter"
)

func init() {
	lox.NO_RETURN_AT_ROOT = false
}

type BusTrackerScript struct {
	statements  []lox.Stmt
	interpreter *lox.Interpreter
}

func NewBusTrackerScript(script string) (bt *BusTrackerScript, err error) {
	scanner := lox.NewScanner(script)
	tokens, err := scanner.ScanTokens()
	if err != nil {
		return
	}

	parser := lox.NewParser(tokens)
	statements, err := parser.Parse()
	if err != nil {
		return
	}

	env := lox.NewEnvironment(nil)
	env.Define("get", &functions.GetFunction{})

	interpreter := lox.NewInterpreter(env)

	resolver := lox.NewResolver(interpreter)
	err = resolver.Resolve(statements...)
	if err != nil {
		return
	}

	return &BusTrackerScript{
		statements:  statements,
		interpreter: interpreter,
	}, nil
}

func (bt *BusTrackerScript) Run() (v interface{}, err error) {
	return bt.interpreter.Interpret(bt.statements)
}
