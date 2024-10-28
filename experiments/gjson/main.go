package main

import (
	"github.com/tidwall/gjson"
)

const json = `{"name":{"first":"Janet","last":"Prichard"},"age":47, "arr": [{"id":1, "name": "one"}, {"id":2, "name": "two"}]}`

func main() {
	value := gjson.Get(json, "arr.#(id==2).name")
	println(value.String())
}

//friends.#(last=="Murphy").first     "Dale"
//friends.#(last=="Murphy")#.first    ["Dale","Jane"]
//friends.#(age>45)#.last             ["Craig","Murphy"]
//friends.#(first%"D*").last          "Murphy"
//friends.#(first!%"D*").last         "Craig
