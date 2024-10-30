package functions

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	lox "github.com/ariyn/lox_interpreter"
	"strings"
)

var cls = lox.NewLoxClass("CrawlData", nil, map[string]lox.Callable{
	"find":   NewBasicFunction("find", 1, find),
	"text":   NewBasicFunction("text", 0, text),
	"attr":   NewBasicFunction("attr", 1, attribute),
	"length": NewBasicFunction("length", 0, length),
	"next":   NewBasicFunction("next", 0, next),
	"parent": NewBasicFunction("parent", 0, parent),
	"html":   NewBasicFunction("html", 0, html),
})

func NewCrawlDataInstance(current string) (*lox.LoxInstance, error) {
	instance := lox.NewLoxInstance(cls)

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(current))
	if err != nil {
		return nil, err
	}

	_ = instance.Set(lox.Token{Lexeme: "doc"}, lox.NewLiteralExpr(doc))

	return instance, nil
}

func NewCrawlDataInstanceWithSelection(doc *goquery.Selection) (*lox.LoxInstance, error) {
	instance := lox.NewLoxInstance(cls)

	_ = instance.Set(lox.Token{Lexeme: "doc"}, lox.NewLiteralExpr(doc))

	return instance, nil
}

func find(doc *goquery.Selection, arguments []any) (v interface{}, err error) {
	if _, ok := arguments[0].(string); !ok {
		err = fmt.Errorf("get() 1st argument need string, but got %v", arguments[0])
		return
	}

	var selector = arguments[0].(string)
	if strings.HasPrefix(selector, "/") {
		selector, err = XpathConverter(selector)

		if err != nil {
			return nil, err
		}
	}

	return doc.Find(selector), nil
}

func text(doc *goquery.Selection, _ []interface{}) (v interface{}, err error) {
	return doc.Text(), nil
}

func attribute(doc *goquery.Selection, arguments []interface{}) (v interface{}, err error) {
	if _, ok := arguments[0].(string); !ok {
		err = fmt.Errorf("get() 1st argument need string, but got %v", arguments[0])
		return
	}

	v, _ = doc.Attr(arguments[0].(string))
	return v, nil
}

func length(doc *goquery.Selection, _ []interface{}) (v interface{}, err error) {
	return doc.Length(), nil
}

func next(doc *goquery.Selection, _ []interface{}) (v interface{}, err error) {
	return doc.Next(), nil
}

func parent(doc *goquery.Selection, _ []interface{}) (v interface{}, err error) {
	return doc.Parent(), nil
}

func html(doc *goquery.Selection, _ []interface{}) (v interface{}, err error) {
	return doc.Html()
}
