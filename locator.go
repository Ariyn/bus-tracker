package bus_tracker

import (
	"fmt"
	lox "github.com/ariyn/lox_interpreter"
	"github.com/playwright-community/playwright-go"
)

const locatorKey = "_Locator"

type locatorFunctionCall func(locator playwright.Locator, arguments []any) (v interface{}, err error)

var _ lox.Callable = (*LocatorFunction)(nil)

type LocatorFunction struct {
	instance *lox.LoxInstance
	arity    int
	call     locatorFunctionCall
	name     string
}

func newLocatorFunction(name string, arity int, call locatorFunctionCall) *LocatorFunction {
	return &LocatorFunction{
		arity: arity,
		call:  call,
		name:  name,
	}
}

func (f LocatorFunction) Call(i *lox.Interpreter, arguments []interface{}) (v interface{}, err error) {
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

func (f LocatorFunction) Arity() int {
	return f.arity
}

func (f LocatorFunction) ToString() string {
	return fmt.Sprintf("<native fn %s>", f.name)
}

func (f LocatorFunction) Bind(instance *lox.LoxInstance) lox.Callable {
	f.instance = instance
	return f
}

func NewLocatorInstance(_locator playwright.Locator) (*lox.LoxInstance, error) {
	instance := lox.NewLoxInstance(lox.NewLoxClass("Locator", nil, map[string]lox.Callable{
		"locator": newLocatorFunction("locator", 1, locator),
		"text": newLocatorFunction("text", 0, func(page playwright.Locator, _ []interface{}) (v interface{}, err error) {
			return page.TextContent()
		}),
		"first": newLocatorFunction("first", 0, func(page playwright.Locator, _ []interface{}) (v interface{}, err error) {
			return NewLocatorInstance(page.First())
		}),
		"last": newLocatorFunction("last", 0, func(page playwright.Locator, _ []interface{}) (v interface{}, err error) {
			return NewLocatorInstance(page.Last())
		}),
		"all": newLocatorFunction("all", 0, func(page playwright.Locator, _ []interface{}) (v interface{}, err error) {
			all, err := page.All()
			if err != nil {
				return nil, err
			}

			instances := make([]*lox.LoxInstance, len(all))
			for i, locator := range all {
				instances[i], err = NewLocatorInstance(locator)
				if err != nil {
					return nil, err
				}
			}

			return lox.ListType{instances}, nil
		}),
	}))

	_ = instance.Set(lox.Token{Lexeme: locatorKey}, lox.NewLiteralExpr(_locator))

	return instance, nil
}

func locator(locator playwright.Locator, arguments []any) (v interface{}, err error) {
	selector, ok := arguments[0].(string)
	if !ok {
		err = fmt.Errorf("get() 1st argument need string, but got %v", arguments[0])
		return
	}
	return NewLocatorInstance(locator.Locator(selector))
}
