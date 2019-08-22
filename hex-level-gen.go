package main

import (
	"encoding/json"
	"log"
	//"math/rand"
	"io/ioutil"
)

func main() {
	jsonMap := make(map[string]interface{})

	jsonMap["target"] = 100
	jsonMap["pushInterval"] = 10
	jsonMap["hexProb"] = 0.7
	jsonMap["bombProb"] = .25
	jsonMap["superProb"] = .05

	jsonMap["row1"] = []string{"blue", "gray", "green", "pink", "red", "yellow", "blue", "gray", "green"}

	outJsonData, err := json.MarshalIndent(jsonMap, "", "    ")
	if err != nil {
		log.Fatal("Error marshalling final level data for output: ", err)
	}

	err = ioutil.WriteFile("1.json", outJsonData, 0644)
	if err != nil {
		log.Fatal("Error writing final json data to level file: ", err)
	}
}
