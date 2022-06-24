package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	fileName := os.Args[1]
	buf, err := ioutil.ReadFile(fileName)
	if err != nil {
		return
	}

	err = Parse(buf)
	if err != nil {
		fmt.Println(err)
		return
	}

}
