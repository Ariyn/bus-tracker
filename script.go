package bus_tracker

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	lox "github.com/ariyn/lox_interpreter"
	"net/http"
	"regexp"
	"strings"
)

var xpathSiblingSelector = regexp.MustCompile(`\[(\d+)\]`)

func xpathConverter(xpath string) (selector string, err error) {
	//xpath := "/html/body/section[3]/div/div/article[5]/div[2]/span[1]/a"
	//convertedSelector := "html > body > section:nth-of-type(3) > div > div > article:nth-of-type(5) > div:nth-of-type(2) > span:first-of-type > a"

	if xpath == "" {
		return "", fmt.Errorf("xpath should not be empty")
	}
	if !strings.HasPrefix(xpath, "/") {
		return "", fmt.Errorf("xpath should start with /")
	}
	xpath = xpath[1:]

	selector = strings.ReplaceAll(xpath, "/", " > ")
	selector = xpathSiblingSelector.ReplaceAllString(selector, ":nth-of-type($1)")

	return selector, nil
}

var _ (lox.Callable) = (*GetFunction)(nil)

type GetFunction struct {
}

func (g GetFunction) Call(_ *lox.Interpreter, arguments []interface{}) (v interface{}, err error) {
	if _, ok := arguments[0].(string); !ok {
		err = fmt.Errorf("get() 1st argument need string, but got %v", arguments[0])
		return
	}
	if _, ok := arguments[1].(string); !ok {
		err = fmt.Errorf("get() 2nd argument need string, but got %v", arguments[1])
		return
	}

	resp, err := http.Get(arguments[0].(string))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var selector = arguments[1].(string)
	if strings.HasPrefix(selector, "/") {
		selector, err = xpathConverter(selector)

		if err != nil {
			return nil, err
		}
	}

	return doc.Find(selector).Text(), nil
}

func (g GetFunction) Arity() int {
	return 2
}

func (g GetFunction) ToString() string {
	return "<native fn Get>"
}

type BusTrackerScript struct {
	interpreter *lox.Interpreter
}

func NewBusTrackerScript(script string) (bt *BusTrackerScript, err error) {
	scanner := lox.NewScanner(script)
	tokens, err := scanner.ScanTokens()
	if err != nil {
		return
	}

	parser := lox.NewParser(tokens)
	statements, err := parser.Parse()
	if err != nil {
		return
	}

	env := lox.NewEnvironment(nil)
	env.Define("get", &GetFunction{})

	interpreter := lox.NewInterpreter(env)

	resolver := lox.NewResolver(interpreter)
	err = resolver.Resolve(statements...)
	if err != nil {
		return
	}

	return &BusTrackerScript{
		interpreter: interpreter,
	}, nil
}
