package bus_tracker

import (
	"encoding/json"
	"fmt"
	"github.com/ariyn/bus-tracker/functions"
	lox "github.com/ariyn/lox_interpreter"
	"io"
	"net/http"
	"regexp"
	"strings"
)

var indexRegexp = regexp.MustCompile(`\[(\d+)\]`)
var contentDispositionRegexp = regexp.MustCompile(`filename(?:\*=UTF-8''|=)(.+)(?:;|$)`)

var _ lox.Callable = (*GetFunction)(nil)

type GetFunction struct {
}

func (g GetFunction) Call(_ *lox.Interpreter, arguments []interface{}) (v interface{}, err error) {
	url, ok := arguments[0].(string)
	if !ok {
		err = fmt.Errorf("get() 1st argument need string, but got %v", arguments[0])
		return
	}

	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "application/json") {
		body, err := convertJsonToXmlString(resp.Body)
		if err != nil {
			return nil, err
		}

		instance, err := functions.NewCrawlDataInstance(body)
		if err != nil {
			return nil, err
		}
		return instance, nil
	}

	if strings.HasPrefix(contentType, "text/html") {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		instance, err := functions.NewCrawlDataInstance(string(body))
		if err != nil {
			return nil, err
		}
		return instance, nil
	}

	if strings.HasPrefix(contentType, "image/") {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		name := resp.Header.Get("Content-Disposition")
		if name == "" {
			tokens := strings.Split(url, "/")
			name = tokens[len(tokens)-1]
		} else {
			matches := contentDispositionRegexp.FindStringSubmatch(name)
			if len(matches) > 1 {
				name = matches[1]
			}
		}

		return NewImageInstance(&Image{Body: body, Url: url, ContentType: resp.Header.Get("Content-Type"), Name: name}), nil
	}

	return nil, fmt.Errorf("Content-Type %s is not supported", contentType)
}

func (g GetFunction) Arity() int {
	return 1
}

func (g GetFunction) ToString() string {
	return "<native fn Get>"
}

func (g GetFunction) Bind(instance *lox.LoxInstance) lox.Callable {
	return g
}

type xmlOption struct {
	key   string
	value interface{}
}

func convertJsonToXmlReader(reader io.Reader) (xmlReader io.ReadCloser, err error) {
	body, err := io.ReadAll(reader)
	if err != nil {
		return
	}

	var v interface{}
	err = json.Unmarshal(body, &v)
	if err != nil {
		return
	}

	xml := convertJsonToXml(&xmlOption{value: v})
	return io.NopCloser(strings.NewReader(xml)), nil
}

func convertJsonToXmlString(reader io.Reader) (xml string, err error) {
	body, err := io.ReadAll(reader)
	if err != nil {
		return
	}

	var v interface{}
	err = json.Unmarshal(body, &v)
	if err != nil {
		return
	}

	xml = convertJsonToXml(&xmlOption{value: v})
	return xml, nil
}

func convertJsonToXml(option *xmlOption) string {
	v := option.value
	k := option.key

	switch v.(type) {
	case string:
		return v.(string)
	case int:
		return fmt.Sprintf("%d", v)
	case float64:
		return fmt.Sprintf("%f", v)
	case bool:
		return fmt.Sprintf("%t", v)
	case map[string]interface{}:
		return convertMapToXml(&xmlOption{value: v.(map[string]interface{})})
	case []interface{}:
		return convertArrayToXml(&xmlOption{key: k, value: v.([]interface{})})
	default:
		return ""
	}
}

func convertMapToXml(option *xmlOption) string {
	m := option.value.(map[string]interface{})
	var xml string
	for k, v := range m {
		xml += fmt.Sprintf("<%s>%s</%s>", k, convertJsonToXml(&xmlOption{value: v, key: k}), k)
	}

	return xml
}

func convertArrayToXml(option *xmlOption) string {
	a := option.value.([]interface{})
	k := option.key
	var xml []string
	for _, v := range a {
		xml = append(xml, convertJsonToXml(&xmlOption{value: v}))
	}
	return strings.Join(xml, fmt.Sprintf("</%s><%s>", k, k))
}
