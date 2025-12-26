package rengine

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/blue-monads/turnix/backend/utils/kosher"
	"github.com/blue-monads/turnix/backend/utils/qq"
	"github.com/tidwall/gjson"
)

type Rule struct {
	ID       string `json:"id"`
	Variable string `json:"variable"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
	ParentID string `json:"parent_id"`
}

// RuleNode represents a rule in the evaluation tree
type RuleNode struct {
	Rule      Rule
	Children  []*RuleNode
	IsGroup   bool
	GroupType string // "AND" or "OR"
}

func RuleEngine(rulestr string, payload []byte) (bool, error) {
	if rulestr == "{}" || rulestr == "" || rulestr == `{"rules":[],"groups":[]}` {
		return true, nil
	}

	ruleData := []Rule{}
	err := json.Unmarshal([]byte(rulestr), &ruleData)
	if err != nil {
		return false, err
	}

	if len(ruleData) == 0 {
		return true, nil
	}

	jsonStr := kosher.Str(payload)

	// Build tree structure from rules
	rootNodes := buildRuleTree(ruleData)

	// Evaluate all root-level rules (they are implicitly ANDed together)
	for _, node := range rootNodes {
		result, err := evaluateNode(node, jsonStr)
		if err != nil {
			return false, err
		}
		if !result {
			return false, nil
		}
	}

	return true, nil
}

// buildRuleTree builds a tree structure from flat rule list
func buildRuleTree(rules []Rule) []*RuleNode {
	nodeMap := make(map[string]*RuleNode)
	var rootNodes []*RuleNode

	// First pass: create all nodes
	for _, rule := range rules {
		node := &RuleNode{
			Rule:     rule,
			Children: []*RuleNode{},
		}

		// Check if this is a logical group
		if rule.Variable == "$logical" && rule.Operator == "group" {
			node.IsGroup = true
			node.GroupType = rule.Value // "AND" or "OR"
		}

		nodeMap[rule.ID] = node
	}

	// Second pass: build parent-child relationships
	for _, rule := range rules {
		node := nodeMap[rule.ID]
		if rule.ParentID == "" {
			rootNodes = append(rootNodes, node)
		} else {
			parent, exists := nodeMap[rule.ParentID]
			if exists {
				parent.Children = append(parent.Children, node)
			} else {
				// Parent not found, treat as root
				rootNodes = append(rootNodes, node)
			}
		}
	}

	return rootNodes
}

// evaluateNode evaluates a rule node and its children
func evaluateNode(node *RuleNode, jsonStr string) (bool, error) {
	if node.IsGroup {
		return evaluateGroup(node, jsonStr)
	}

	return evaluateRule(node.Rule, jsonStr)
}

// evaluateGroup evaluates a logical group (AND/OR)
func evaluateGroup(group *RuleNode, jsonStr string) (bool, error) {
	if len(group.Children) == 0 {
		return true, nil
	}

	if group.GroupType == "OR" {
		// OR: at least one child must be true
		for _, child := range group.Children {
			result, err := evaluateNode(child, jsonStr)
			if err != nil {
				return false, err
			}
			if result {
				return true, nil
			}
		}
		return false, nil
	} else {
		// AND (default): all children must be true
		for _, child := range group.Children {
			result, err := evaluateNode(child, jsonStr)
			if err != nil {
				return false, err
			}
			if !result {
				return false, nil
			}
		}
		return true, nil
	}
}

// evaluateRule evaluates a single rule against the JSON payload
func evaluateRule(rule Rule, jsonStr string) (bool, error) {
	// Get the value from JSON using the variable path
	value := gjson.Get(jsonStr, rule.Variable)

	// If variable doesn't exist, treat as empty string
	actualValue := ""
	if value.Exists() {
		actualValue = value.String()
	}
	expectedValue := rule.Value

	switch rule.Operator {
	case "equal_to":
		return actualValue == expectedValue, nil

	case "not_equal_to":
		return actualValue != expectedValue, nil

	case "contains":
		return strings.Contains(actualValue, expectedValue), nil

	case "not_contains":
		return !strings.Contains(actualValue, expectedValue), nil

	case "greater_than":
		return compareValues(actualValue, expectedValue) > 0, nil

	case "less_than":
		return compareValues(actualValue, expectedValue) < 0, nil

	case "greater_than_or_equal":
		comp := compareValues(actualValue, expectedValue)
		return comp >= 0, nil

	case "less_than_or_equal":
		comp := compareValues(actualValue, expectedValue)
		return comp <= 0, nil

	case "before":
		return compareDates(actualValue, expectedValue) < 0, nil

	case "after":
		return compareDates(actualValue, expectedValue) > 0, nil

	default:
		qq.Println("RuleEngine: unknown operator", rule.Operator)
		return false, nil
	}
}

// compareValues compares two values, trying numeric comparison first, then string
func compareValues(a, b string) int {
	// Try numeric comparison
	aNum, aErr := strconv.ParseFloat(a, 64)
	bNum, bErr := strconv.ParseFloat(b, 64)
	if aErr == nil && bErr == nil {
		if aNum > bNum {
			return 1
		} else if aNum < bNum {
			return -1
		}
		return 0
	}

	// Fall back to string comparison
	if a > b {
		return 1
	} else if a < b {
		return -1
	}
	return 0
}

// compareDates compares two date strings
func compareDates(a, b string) int {
	// Try various date formats
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05",
		"2006-01-02",
		time.RFC1123,
		time.RFC1123Z,
	}

	var aTime, bTime time.Time
	var aErr, bErr error

	for _, format := range formats {
		aTime, aErr = time.Parse(format, a)
		if aErr == nil {
			break
		}
	}

	for _, format := range formats {
		bTime, bErr = time.Parse(format, b)
		if bErr == nil {
			break
		}
	}

	// If both parsed successfully, compare
	if aErr == nil && bErr == nil {
		if aTime.After(bTime) {
			return 1
		} else if aTime.Before(bTime) {
			return -1
		}
		return 0
	}

	// If parsing failed, fall back to string comparison
	return compareValues(a, b)
}
