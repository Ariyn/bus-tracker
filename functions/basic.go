package functions

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	lox "github.com/ariyn/lox_interpreter"
)

type BasicFunctionCall func(doc *goquery.Selection, arguments []any) (v interface{}, err error)

var _ (lox.Callable) = (*BasicFunction)(nil)

type BasicFunction struct {
	instance *lox.LoxInstance
	arity    int
	call     BasicFunctionCall
	name     string
}

func NewBasicFunction(name string, arity int, call BasicFunctionCall) *BasicFunction {
	return &BasicFunction{
		arity: arity,
		call:  call,
		name:  name,
	}
}

func (bf BasicFunction) Call(i *lox.Interpreter, arguments []interface{}) (v interface{}, err error) {
	data, err := bf.instance.Get(lox.Token{Lexeme: "doc"})
	if err != nil {
		return
	}

	doc, isDocument := data.(*goquery.Document)
	if isDocument {
		v, err = bf.call(doc.Selection, arguments)
	}

	selection, isSelection := data.(*goquery.Selection)
	if isSelection {
		v, err = bf.call(selection, arguments)
	}

	if !isDocument && !isSelection {
		return nil, fmt.Errorf("doc is not Document")
	}

	if err != nil {
		return nil, err
	}

	if _, ok := v.(string); ok {
		return v, nil
	}

	instance, err := NewCrawlDataInstanceWithSelection(v.(*goquery.Selection))
	return instance, err
}

func (bf BasicFunction) Arity() int {
	return bf.arity
}

func (bf BasicFunction) ToString() string {
	return fmt.Sprintf("<native fn %s>", bf.name)
}

func (bf BasicFunction) Bind(instance *lox.LoxInstance) lox.Callable {
	bf.instance = instance
	return bf
}
