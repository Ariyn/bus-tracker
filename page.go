package bus_tracker

import (
	"fmt"
	lox "github.com/ariyn/lox_interpreter"
	"github.com/playwright-community/playwright-go"
)

func NewPageInstance(page playwright.Page) (*lox.LoxInstance, error) {
	instance := lox.NewLoxInstance(
		lox.NewLoxClass("Page", nil, map[string]lox.Callable{
			"locator": newFunction("locator", 1, func(page playwright.Page, arguments []any) (v interface{}, err error) {
				selector, ok := arguments[0].(string)
				if !ok {
					err = fmt.Errorf("get() 1st argument need string, but got %v", arguments[0])
					return
				}

				return NewLocatorInstance(page.Locator(selector))
			}),
			"screenshot": newFunction("image", 0, func(page playwright.Page, arguments []any) (v interface{}, err error) {
				image, err := page.Screenshot(playwright.PageScreenshotOptions{
					FullPage: playwright.Bool(true),
				})

				if err != nil {
					return nil, fmt.Errorf("could not take screenshot: %v", err)
				}

				return NewImageInstance(&Image{
					Url:         "",
					Body:        image,
					Name:        "screenshot.png",
					ContentType: "image/png",
				}), nil
			}),
		}))

	_ = instance.Set(lox.Token{Lexeme: "page"}, lox.NewLiteralExpr(page))

	return instance, nil
}

type pageFunctionCall func(doc playwright.Page, arguments []any) (v interface{}, err error)

var _ lox.Callable = (*PageFunction)(nil)

type PageFunction struct {
	instance *lox.LoxInstance
	arity    int
	call     pageFunctionCall
	name     string
}

func newFunction(name string, arity int, call pageFunctionCall) *PageFunction {
	return &PageFunction{
		arity: arity,
		call:  call,
		name:  name,
	}
}

func (f PageFunction) Call(i *lox.Interpreter, arguments []interface{}) (v interface{}, err error) {
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

func (f PageFunction) Arity() int {
	return f.arity
}

func (f PageFunction) ToString() string {
	return fmt.Sprintf("<native fn %s>", f.name)
}

func (f PageFunction) Bind(instance *lox.LoxInstance) lox.Callable {
	f.instance = instance
	return f
}
