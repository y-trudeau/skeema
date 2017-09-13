// Package mysqlgoutils is a set a functions needed to extent the way skeema works

package mysqlgoutils

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// SplitHostOptionalPortAndSchem takes an address string containing a hostname,
// ipv4 addr, or ipv6 addr; *optionally* followed by a colon and port number
// optionally followed by a pipe '|' and a schema name. It  splits the hostname
// portion from the port portion and the schema portion and returns them
// separately. If no port was present, 0 will be returned for that portion.
// If no schema was present, '' will be returned for that portion.
// If hostaddr contains an ipv6 address, the IP address portion must be
// wrapped in brackets on input, and the brackets will still be present on
// output.
func SplitHostOptionalPortAndSchema(hostaddr string) (string, int, string, error) {
	if len(hostaddr) == 0 {
		return "", 0, errors.New("Cannot parse blank host address")
	}

	// ipv6 without port, or ipv4 or hostname without port
	if (hostaddr[0] == '[' && hostaddr[len(hostaddr)-1] == ']') || len(strings.Split(hostaddr, "|")) == 1 {
		return hostaddr, 0, "", nil
	}

	var schema string
	schema = strings.Split(hostaddr, "|")[1]

	host, port, err := net.SplitHostOptionalPort(strings.Split(hostaddr, "|")[0])
	if err != nil {
		return "", 0, "", err
	}

	return host, port, schema, nil
}
