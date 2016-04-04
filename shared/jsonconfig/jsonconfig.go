package jsonconfig

import (
	"io"
	"io/ioutil"
	"log"
	"os"
)

type Parser interface {
	ParseJSON([]byte) error
}

func Load(configFile string, parser Parser) {
	var err error
	var input = io.ReadCloser(os.Stdin)
	if input, err = os.Open(configFile); err != nil {
		log.Fatalln(err)
	}

	// Read the config file
	jsonBytes, err := ioutil.ReadAll(input)
	input.Close()
	if err != nil {
		log.Fatalln(err)
	}

	if err := parser.ParseJSON(jsonBytes); err != nil {
		log.Fatalln("Could not parse %q: %v", configFile, err)
	}
}
