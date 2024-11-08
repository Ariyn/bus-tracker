package bus_tracker

import (
	"context"
	"fmt"
	lox "github.com/ariyn/lox_interpreter"
	"github.com/playwright-community/playwright-go"
	"io"
	"log"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"
)

var initScript = ""

func init() {
	f, err := os.Open(os.Getenv("PLAYWRIGHT_BROWSER_INIT_SCRIPT_PATH"))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}

	initScript = string(b)
}

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
		SkipInstallBrowsers: false,
		Stdout:              os.Stdout,
		Stderr:              os.Stderr,
	})
	if err != nil {
		return nil, fmt.Errorf("could not start playwright: %v", err)
	}

	_browser, err := pw.Firefox.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false),
		Args:     []string{"--incognito"},
	})
	if err != nil {
		return nil, fmt.Errorf("could not launch browser: %v", err)
	}

	_page, err := _browser.NewPage(playwright.BrowserNewPageOptions{
		UserAgent: playwright.String(userAgent.(string)),
		Locale:    playwright.String("ko-KR"),
		ExtraHttpHeaders: map[string]string{
			"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
			"Accept-Encoding":           "gzip, deflate, br, zstd",
			"Accept-Language":           "ko-KR,ko;q=0.9",
			"Cache-Control":             "no-cache",
			"Cookie":                    "",
			"Pragma":                    "no-cache",
			"Sec-Ch-Ua":                 `"Google Chrome";v="123", "Not:A-Brand";v="8", "Chromium";v="123"`,
			"Sec-Fetch-Dest":            "document",
			"Sec-Fetch-Mode":            "navigate",
			"Sec-Fetch-Site":            "none",
			"Sec-Fetch-User":            "?1",
			"Upgrade-Insecure-Requests": "1",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("could not create page: %v", err)
	}

	_ = _page.SetViewportSize(1920, 1080)
	evaluatedResult, err := _page.Evaluate(``)

	if err != nil {
		return
	}

	log.Println("evaluated", evaluatedResult)

	//waitCtx, cancel := context.WithCancel(context.Background())
	//defer cancel()

	//move := moveMouseRandom(waitCtx, _page)

	if _, err = _page.Goto(url); err != nil {
		return nil, fmt.Errorf("could not goto: %v", err)
	}
	//go move()

	err = _page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{State: playwright.LoadStateDomcontentloaded})
	if err != nil {
		return nil, fmt.Errorf("could not wait for load state: %v", err)
	}

	//cancel()

	browsers, err := i.Globals.Get(lox.Token{Lexeme: "_browsers"})
	if err != nil {
		browsers = make([]playwright.Browser, 0)
		browsers = append(browsers.([]playwright.Browser), _browser)

		i.Globals.Define("_browsers", browsers)
	} else {
		browsers = append(browsers.([]playwright.Browser), _browser)
		i.Globals.Assign(lox.Token{Lexeme: "_browsers"}, browsers)
	}

	return NewPageInstance(_page)
}

func (f BrowserGetFunction) Arity() int {
	return 1
}

func (f BrowserGetFunction) ToString() string {
	return "<native fn Browser>"
}

func moveMouseRandom(ctx context.Context, page playwright.Page) (move func()) {
	mouse := page.Mouse()
	currentX, currentY := 130.0, 250.0
	screenWidth, screenHeight := 1920, 1080

	_ = mouse.Move(currentX, currentY)

	return func() {
		nextTargetX, nextTargetY := currentX, currentY

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			if nextTargetX == currentX && nextTargetY == currentY {
				nextTargetX, nextTargetY = float64(rand.Int()%screenWidth), float64(rand.Int()%screenHeight)
			}

			deltaX, deltaY := 0.0, 0.0
			if nextTargetX-currentX != 0 {
				deltaX = (nextTargetX - currentX) / math.Abs(nextTargetX-currentX)
			}
			if nextTargetY-currentY != 0 {
				deltaY = 1
			}
			_ = mouse.Move(deltaX, deltaY)

			time.Sleep(time.Millisecond * time.Duration(rand.Intn(30)))
		}
	}
}
