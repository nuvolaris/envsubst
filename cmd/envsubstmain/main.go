// envsubst command line tool
package envsubstmain

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/a8m/envsubst/parse"
)

var (
	input    = flag.String("i", "", "")
	output   = flag.String("o", "", "")
	noDigit  = flag.Bool("no-digit", false, "")
	noUnset  = flag.Bool("no-unset", false, "")
	noEmpty  = flag.Bool("no-empty", false, "")
	failFast = flag.Bool("fail-fast", false, "")
)

var usage = `Usage: envsubst [options...] <input>
Options:
  -i         Specify file input, otherwise read from stdin.
  -o         Specify file output. If none is specified, write to stdout.
  -no-digit  Do not replace variables starting with a digit. e.g. $1 and ${1}
  -no-unset  Fail if a variable is not set.
  -no-empty  Fail if a variable is set but empty.
  -fail-fast Fail on first error otherwise display all failures if restrictions are set.
`

func EnvsubstMain() error {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, fmt.Sprintf(usage))
	}
	flag.Parse()
	var reader *bufio.Reader
	if *input != "" {
		file, err := os.Open(*input)
		if err != nil {
			return usageAndExit(fmt.Sprintf("Error to open file input: %s.", *input))
		}
		defer file.Close()
		reader = bufio.NewReader(file)
	} else {
		stat, err := os.Stdin.Stat()
		if err != nil || (stat.Mode()&os.ModeCharDevice) != 0 {
			return usageAndExit("Error: no input received.")
		}
		reader = bufio.NewReader(os.Stdin)
	}
	// Collect input data.
	var data string
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				data += line
				break
			}
			return usageAndExit("Failed to read input.")
		}
		data += line
	}
	var (
		err  error
		file *os.File
	)
	if *output != "" {
		file, err = os.Create(*output)
		if err != nil {
			return usageAndExit("Error to create the wanted output file.")
		}
	} else {
		file = os.Stdout
	}
	// Parse input string
	parserMode := parse.AllErrors
	if *failFast {
		parserMode = parse.Quick
	}
	restrictions := &parse.Restrictions{*noUnset, *noEmpty, *noDigit}
	result, err := (&parse.Parser{Name: "string", Env: os.Environ(), Restrict: restrictions, Mode: parserMode}).Parse(data)
	if err != nil {
		return err
	}
	if _, err := file.WriteString(result); err != nil {
		filename := *output
		if filename == "" {
			filename = "STDOUT"
		}
		return usageAndExit(fmt.Sprintf("Error writing output to: %s.", filename))
	}

	return nil
}

func usageAndExit(msg string) error {
	// if msg != "" {
	// 	fmt.Fprintf(os.Stderr, msg)
	// 	fmt.Fprintf(os.Stderr, "\n\n")
	// }
	flag.Usage()
	fmt.Fprintf(os.Stderr, "\n")
	return errors.New(msg)
}
