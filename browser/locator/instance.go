package locator

import (
	"fmt"
	lox "github.com/ariyn/lox_interpreter"
	"github.com/playwright-community/playwright-go"
)

const locatorKey = "_Locator"

var methods = map[string]lox.Callable{}

func init() {
	methods = map[string]lox.Callable{
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
	}
}

func NewLocatorInstance(_locator playwright.Locator) (*lox.LoxInstance, error) {
	instance := lox.NewLoxInstance(lox.NewLoxClass("Locator", nil, methods))

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
