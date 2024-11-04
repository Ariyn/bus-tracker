package page

import (
	"fmt"
	lox "github.com/ariyn/lox_interpreter"
	"github.com/playwright-community/playwright-go"
)

type functionCall func(doc playwright.Page, arguments []any) (v interface{}, err error)

var _ lox.Callable = (*Function)(nil)

type Function struct {
	instance *lox.LoxInstance
	arity    int
	call     functionCall
	name     string
}

func newFunction(name string, arity int, call functionCall) *Function {
	return &Function{
		arity: arity,
		call:  call,
		name:  name,
	}
}

func (f Function) Call(i *lox.Interpreter, arguments []interface{}) (v interface{}, err error) {
	page, err := f.instance.Get(lox.Token{Lexeme: "page"})
	if err != nil {
		return
	}

	_, isDocument := page.(playwright.Page)
	if !isDocument {
		return nil, fmt.Errorf("is not Document")
	}

	v, err = f.call(page.(playwright.Page), arguments)

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
