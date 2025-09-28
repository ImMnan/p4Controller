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
	"os"
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
func ServersRead() ([]ServerJSON, error) {

	p4PortMaster := os.Getenv("P4PORT")
	// Run the p4 -Mj -ztag servers command
	infoCmd := exec.Command("p4", "-p", p4PortMaster, "info")
	outputInfo, err := infoCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run p4 command: %w", err)
	}
	// Check if the output is empty
	if len(outputInfo) == 0 {
		return nil, fmt.Errorf("p4 command returned no output")
	}

	fmt.Printf("Executing command: %s\n", infoCmd.String())

	cmd := exec.Command("p4", "-p", p4PortMaster, "-Mj", "-ztag", "servers")
	fmt.Printf("Executing command: %s\n", cmd.String())

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run p4 command: %w", err)
	}

	// Check if the output is empty
	if len(output) == 0 {
		return nil, fmt.Errorf("p4 command returned no output")
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

	// Print the parsed servers to the terminal
	for i, s := range servers {
		fmt.Printf("Server %d: %+v\n", i+1, s)
	}

	return servers, nil
}
