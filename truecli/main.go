package main

import (
	"bytes"
	"fmt"
	"log"
	"os"

	"github.com/davecgh/go-spew/spew"

	"github.com/unidoc/unipdf/v3/common"

	"github.com/gunnsth/unitype"
)

func main() {
	spew.Dump(os.Args)
	common.SetLogger(common.NewConsoleLogger(common.LogLevelDebug))

	if len(os.Args) < 2 {
		fmt.Printf("Missing argument\n")
		return
	}
	fmt.Println("blah")

	tfnt, err := unitype.ParseFile(os.Args[1])
	if err != nil {
		log.Fatalf("Error: %+v", err)
	}

	fmt.Printf("tfnt----\n")
	fmt.Printf("%s\n", tfnt.String())

	var buf bytes.Buffer
	err = tfnt.Write(&buf)
	if err != nil {
		fmt.Printf("Error writing: %+v\n", err)
		return
	}

	err = unitype.ValidateBytes(buf.Bytes())
	if err != nil {
		fmt.Printf("Invalid font: %+v\n", err)
		panic(err)
	} else {
		fmt.Printf("Font is valid\n")
	}

	err = unitype.ValidateFile(os.Args[1])
	if err != nil {
		fmt.Printf("Invalid font: %+v\n", err)
		panic(err)
	} else {
		fmt.Printf("Font is valid\n")
	}

	fmt.Println("---123")
	// Try subsetting font.
	subfnt, err := tfnt.SubsetSimple(256)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Subset font: %s\n", subfnt.String())

	buf.Reset()
	err = subfnt.Write(&buf)
	if err != nil {
		fmt.Printf("Failed writing: %+v\n", err)
		panic(err)
	}
	fmt.Printf("Subset font length: %d\n", buf.Len())
	err = unitype.ValidateBytes(buf.Bytes())
	if err != nil {
		fmt.Printf("Invalid subfnt: %+v\n", err)
		panic(err)
	} else {
		fmt.Printf("subset font is valid\n")
	}

	err = subfnt.WriteFile("subset.ttf")
	if err != nil {
		panic(err)
	}

}
