package p4c

// This package is resonsible for working with the p4d server with p4 command
// This function will READ the p4 servers list.
// The command is p4 -Mj -ztag servers
// The output is in JSON format
// We will parse the JSON output and return an output in Config struct
// {"Address":"","Description":"Created by bruno.\n","Name":"","Options":"nomandatory","ServerID":"commit","Services":"standard","Type":"server"}

// Name : Config.P4CSpec key

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
)

// ServerJSON represents the structure of a server entry from 'p4 -Mj -ztag servers' JSON output
type ServerJSON struct {
	Address     string `json:"Address"`
	Description string `json:"Description"`
	Name        string `json:"Name"`
	Options     string `json:"Options"`
	ServerID    string `json:"ServerID"`
	Services    string `json:"Services"`
	Type        string `json:"Type"`
}

// serversRead runs 'p4 -Mj -ztag servers', parses the JSON output, and returns a slice of ServerJSON
func serversRead() ([]ServerJSON, error) {
	// Run the p4 -Mj -ztag servers command
	cmd := exec.Command("p4", "-Mj", "-ztag", "servers")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run p4 command: %w", err)
	}

	// Parse the JSON output (one JSON object per line)
	dec := json.NewDecoder(bytes.NewReader(output))
	var servers []ServerJSON
	for dec.More() {
		var server ServerJSON
		if err := dec.Decode(&server); err != nil {
			return nil, fmt.Errorf("failed to decode JSON: %w", err)
		}
		servers = append(servers, server)
	}

	return servers, nil
}
