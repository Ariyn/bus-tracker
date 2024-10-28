package main

import (
	"github.com/joho/godotenv"
	"github.com/tidwall/gjson"
	"io"
	"net/http"
	"os"
)

func main() {
	godotenv.Load()

	url := os.Getenv("PROTOTYPING_TARGET_URL")
	script := os.Getenv("PROTOTYPING_GJSON_SCRIPT")

	println(url, script)
	response, _ := http.Get(url)
	defer response.Body.Close()

	b, _ := io.ReadAll(response.Body)
	println(string(b))

	value := gjson.Get(string(b), script)
	println(value.String())
}
