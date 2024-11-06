package bus_tracker

import (
	"fmt"
	lox "github.com/ariyn/lox_interpreter"
	"github.com/playwright-community/playwright-go"
	"strings"
)

var _ lox.Callable = (*BrowserGetFunction)(nil)

type BrowserGetFunction struct {
}

func (f BrowserGetFunction) Bind(instance *lox.LoxInstance) lox.Callable {
	return f
}

func (f BrowserGetFunction) Call(i *lox.Interpreter, arguments []interface{}) (v interface{}, err error) {
	userAgent, err := i.Globals.Get(lox.Token{Lexeme: "$user-agent"})
	if err != nil {
		if strings.Index(err.Error(), "Undefined variable") != -1 {
			userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3"
		} else {
			return nil, err
		}
	}

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

	_page, err := _browser.NewPage(playwright.BrowserNewPageOptions{
		UserAgent: playwright.String(userAgent.(string)),
	})
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

	return NewPageInstance(_page)
}

func (f BrowserGetFunction) Arity() int {
	return 1
}

func (f BrowserGetFunction) ToString() string {
	return "<native fn Browser>"
}
