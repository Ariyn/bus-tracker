package page

import (
	"fmt"
	locator2 "github.com/ariyn/bus-tracker/browser/locator"
	lox "github.com/ariyn/lox_interpreter"
	"github.com/playwright-community/playwright-go"
)

func NewInstance(page playwright.Page) (*lox.LoxInstance, error) {
	instance := lox.NewLoxInstance(
		lox.NewLoxClass("Page", nil, map[string]lox.Callable{
			"locator": newFunction("locator", 1, locator),
		}))

	_ = instance.Set(lox.Token{Lexeme: "page"}, lox.NewLiteralExpr(page))

	return instance, nil
}

func locator(page playwright.Page, arguments []any) (v interface{}, err error) {
	selector, ok := arguments[0].(string)
	if !ok {
		err = fmt.Errorf("get() 1st argument need string, but got %v", arguments[0])
		return
	}

	return locator2.NewLocatorInstance(page.Locator(selector))
}
