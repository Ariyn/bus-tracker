package locator

import (
	"fmt"
	lox "github.com/ariyn/lox_interpreter"
	"github.com/playwright-community/playwright-go"
)

type functionCall func(locator playwright.Locator, arguments []any) (v interface{}, err error)

var _ lox.Callable = (*Function)(nil)

type Function struct {
	instance *lox.LoxInstance
	arity    int
	call     functionCall
	name     string
}

func newLocatorFunction(name string, arity int, call functionCall) *Function {
	return &Function{
		arity: arity,
		call:  call,
		name:  name,
	}
}

func (f Function) Call(i *lox.Interpreter, arguments []interface{}) (v interface{}, err error) {
	page, err := f.instance.Get(lox.Token{Lexeme: locatorKey})
	if err != nil {
		return
	}

	_, isLocator := page.(playwright.Locator)
	if !isLocator {
		return nil, fmt.Errorf("is not Locator")
	}

	v, err = f.call(page.(playwright.Locator), arguments)

	return v, err
}

func (f Function) Arity() int {
	return f.arity
}

func (f Function) ToString() string {
	return fmt.Sprintf("<native fn %s>", f.name)
}

func (f Function) Bind(instance *lox.LoxInstance) lox.Callable {
	f.instance = instance
	return f
}
