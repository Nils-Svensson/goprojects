package audit

import (
	"encoding/json"
	"fmt"
	"os"

	"goprojects/findings"

	"gopkg.in/yaml.v3"
)

func OutputFindingsAsJSON(findingsList []findings.Finding, filename string) error {
	data, err := json.MarshalIndent(findingsList, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal findings: %w", err)
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write JSON file: %w", err)
	}

	return nil
}

func OutputFindingsAsYAML(findingsList []findings.Finding, filename string) error {
	data, err := yaml.Marshal(findingsList)
	if err != nil {
		return fmt.Errorf("failed to marshal findings: %w", err)
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write YAML file: %w", err)
	}

	return nil
}
