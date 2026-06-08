package engine

import (
	"regexp"
	"strings"
)

// extractEntitiesLocal performs fast regex-based entity extraction on the Go side.
// This mirrors the Rust router's extract_entities() logic so we can check L1b cache
// without a gRPC round-trip. Runs in <1ms.
func extractEntitiesLocal(text string) map[string]string {
	entities := make(map[string]string)
	lower := strings.ToLower(text)

	// --- Money Extractor ---
	moneyRe := regexp.MustCompile(`(?i)\$\s*(\d[\d,]*(?:\.\d+)?)|(\d[\d,]*(?:\.\d+)?)\s*(?:dollars?|usd)`)
	if matches := moneyRe.FindStringSubmatch(text); len(matches) > 0 {
		amount := matches[1]
		if amount == "" {
			amount = matches[2]
		}
		amount = strings.ReplaceAll(amount, ",", "")
		if amount != "" {
			entities["amount"] = amount
		}
	}

	// --- Directional Transfer Detection ---
	transferRe := regexp.MustCompile(`(?i)(?:from|source)\s+(?:account\s*#?\s*)?(\d+)\s+(?:to|into|→)\s+(?:account\s*#?\s*)?(\d+)`)
	if matches := transferRe.FindStringSubmatch(text); len(matches) > 0 {
		entities["source_account"] = matches[1]
		entities["target_account"] = matches[2]
	} else {
		// Fallback: generic account/ID extraction
		idRe := regexp.MustCompile(`(?i)(?:account|acct|id|invoice|order|customer|user)\s*#?\s*(\d+)`)
		idMatches := idRe.FindAllStringSubmatch(text, -1)
		if len(idMatches) >= 2 {
			entities["entity_id_1"] = idMatches[0][1]
			entities["entity_id_2"] = idMatches[1][1]
		} else if len(idMatches) == 1 {
			entities["entity_id"] = idMatches[0][1]
		}
	}

	// --- Action Extractor ---
	actions := []struct {
		key      string
		keywords []string
	}{
		{"transfer", []string{"transfer", "send", "move", "wire"}},
		{"balance", []string{"balance", "owe", "owes", "outstanding", "dues", "due amount"}},
		{"lookup", []string{"lookup", "look up", "find", "search", "get", "fetch", "show", "check"}},
		{"create", []string{"create", "add", "new", "register"}},
		{"delete", []string{"delete", "remove", "cancel"}},
		{"update", []string{"update", "edit", "modify", "change"}},
	}
	for _, a := range actions {
		for _, kw := range a.keywords {
			if strings.Contains(lower, kw) {
				entities["action"] = a.key
				goto actionDone
			}
		}
	}
actionDone:

	// --- Customer/Name Extractor ---
	// Match "customer X" or "client X" directly; "for" is too generic so require "for customer/client"
	nameRe := regexp.MustCompile(`(?i)(?:customer|client)\s+([A-Z][a-zA-Z]+(?:\s+[A-Z][a-zA-Z]+)?)`)
	if matches := nameRe.FindStringSubmatch(text); len(matches) > 0 {
		entities["customer"] = matches[1]
	}

	// --- Date Extractor ---
	dateRe := regexp.MustCompile(`(\d{4}-\d{2}-\d{2}|\d{1,2}/\d{1,2}/\d{4})`)
	if matches := dateRe.FindStringSubmatch(text); len(matches) > 0 {
		entities["date"] = matches[1]
	}

	return entities
}
