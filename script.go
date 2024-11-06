package bus_tracker

import (
	"fmt"
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

func NewBusTrackerScript(script string, envVar map[string]string) (bt *BusTrackerScript, err error) {
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
	env.Define("get", &GetFunction{})
	env.Define("browser", &BrowserGetFunction{})
	env.Define("number", &NumberFunction{})

	for k, v := range envVar {
		env.Define(k, v)
	}

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
	v, err = bt.interpreter.Interpret(bt.statements)
	if err != nil {
		return
	}

	if instance, ok := v.(*lox.LoxInstance); ok {
		if instance.ToString() == "<inst Image>" {
			return instance.Get(lox.Token{Lexeme: "_image"})
		}
	}

	return v, nil
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
