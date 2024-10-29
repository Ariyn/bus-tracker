package functions

import (
	"fmt"
	"regexp"
	"strings"
)

var xpathSiblingSelector = regexp.MustCompile(`\[(\d+)\]`)

func XpathConverter(xpath string) (selector string, err error) {
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
