package functions

import (
	"encoding/json"
	"fmt"
	lox "github.com/ariyn/lox_interpreter"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
)

var indexRegexp = regexp.MustCompile(`\[(\d+)\]`)

var _ (lox.Callable) = (*GetFunction)(nil)

type GetFunction struct {
}

func (g GetFunction) Call(_ *lox.Interpreter, arguments []interface{}) (v interface{}, err error) {
	if _, ok := arguments[0].(string); !ok {
		err = fmt.Errorf("get() 1st argument need string, but got %v", arguments[0])
		return
	}

	resp, err := http.Get(arguments[0].(string))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "application/json") {
		resp.Body, err = convertJsonToXmlReader(resp.Body)
		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	instance, err := NewCrawlDataInstance(string(body))
	if err != nil {
		return nil, err
	}

	return instance, nil
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
	log.Println(xml)
	return io.NopCloser(strings.NewReader(xml)), nil
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
		log.Println(k, xml)
		//xml += convertJsonToXml(&xmlOption{value: v, key: k})
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
