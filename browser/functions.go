package browser

import (
	"fmt"
	"github.com/ariyn/bus-tracker/browser/page"
	lox "github.com/ariyn/lox_interpreter"
	"github.com/playwright-community/playwright-go"
)

var _ lox.Callable = (*GetFunction)(nil)

type GetFunction struct {
}

func (f GetFunction) Bind(instance *lox.LoxInstance) lox.Callable {
	return f
}

func (f GetFunction) Call(_ *lox.Interpreter, arguments []interface{}) (v interface{}, err error) {
	url, ok := arguments[0].(string)
	if !ok {
		return nil, fmt.Errorf("playwright() 1st argument need string, but got %v", arguments[0])
	}

	pw, err := playwright.Run(&playwright.RunOptions{
		Browsers: []string{"firefox"},
	})
	if err != nil {
		return nil, fmt.Errorf("could not start playwright: %v", err)
	}

	_browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
		Args:     []string{"--incognito"},
	})
	if err != nil {
		return nil, fmt.Errorf("could not launch browser: %v", err)
	}

	_page, err := _browser.NewPage()
	if err != nil {
		return nil, fmt.Errorf("could not create page: %v", err)
	}

	if _, err = _page.Goto(url); err != nil {
		return nil, fmt.Errorf("could not goto: %v", err)
	}

	err = _page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{State: playwright.LoadStateDomcontentloaded})
	if err != nil {
		return nil, fmt.Errorf("could not wait for load state: %v", err)
	}

	return page.NewInstance(_page)
}

func (f GetFunction) Arity() int {
	return 1
}

func (f GetFunction) ToString() string {
	return "<native fn Browser>"
}
