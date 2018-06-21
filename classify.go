package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"encoding/json"
)

func main() {
	cardsBytes, err := ioutil.ReadFile("cards.json")
	if err != nil {
		log.Fatal(err)
	}
	type CardsTopLevel struct {
		ImageUrlTemplate string `json:"imageUrlTemplate"`
		Data []map[string]interface{} `json:"data"`
	}
	var cards CardsTopLevel
	err = json.Unmarshal(cardsBytes, &cards)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Success:", cards.ImageUrlTemplate)
	for k, v := range cards.Data[0] {
		switch vv := v.(type) {
		case nil:
			fmt.Println(k, "is nil")
		case bool:
			fmt.Println(k, "is boolean", vv)
		case string:
			fmt.Println(k, "is string", vv)
		case float64:
			fmt.Println(k, "is float64", vv)
		case []interface{}:
			fmt.Println(k, "is an array")
		default:
			fmt.Println(k, "is of a type I don't know how to handle")
		}
	}
}
