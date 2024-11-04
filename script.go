package bus_tracker

import (
	"fmt"
	"github.com/ariyn/bus-tracker/browser"
	"github.com/ariyn/bus-tracker/functions"
	lox "github.com/ariyn/lox_interpreter"
	"strconv"
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
	env.Define("browser", &browser.GetFunction{})
	env.Define("number", &NumberFunction{})

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

var _ lox.Callable = (*NumberFunction)(nil)

type NumberFunction struct {
}

func (n NumberFunction) Call(interpreter *lox.Interpreter, arguments []interface{}) (interface{}, error) {
	arg, ok := arguments[0].(string)
	if !ok {
		if _, ok := arguments[0].(float64); !ok {
			return nil, fmt.Errorf("number() argument must be string or number")
		}

		return arguments[0], nil
	}

	return strconv.ParseFloat(arg, 64)
}

func (n NumberFunction) Arity() int {
	return 1
}

func (n NumberFunction) ToString() string {
	return "<native fn>"
}

func (n NumberFunction) Bind(instance *lox.LoxInstance) lox.Callable {
	return n
}
