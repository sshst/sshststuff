package tlslog

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

var Writer io.Writer

func init() {
	if path, found := os.LookupEnv("SSLKEYLOGFILE"); found {
		w, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND, 0600)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening SSLKEYLOGFILE: %+v", err)
		} else {
			Writer = w
		}
	} else {
		Writer = ioutil.Discard
	}
}
