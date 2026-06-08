package tools

import (
	"fmt"
	"strings"
)

// SupportedConnectorTypes lists connector types with a working gateway implementation.
var SupportedConnectorTypes = []ToolConnectorType{
	ConnectorCore,
	ConnectorMCP,
	ConnectorREST,
	ConnectorSQL,
}

func IsSupportedConnectorType(t string) bool {
	normalized := strings.ToLower(strings.TrimSpace(t))
	for _, supported := range SupportedConnectorTypes {
		if string(supported) == normalized {
			return true
		}
	}
	return false
}

func SupportedConnectorTypesString() string {
	parts := make([]string, len(SupportedConnectorTypes))
	for i, t := range SupportedConnectorTypes {
		parts[i] = string(t)
	}
	return strings.Join(parts, ", ")
}

func UnsupportedConnectorError(got string) error {
	return fmt.Errorf("unsupported connector_type %q — supported: %s", got, SupportedConnectorTypesString())
}
