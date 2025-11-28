package eventhub

import (
	"encoding/json"
	"testing"
)

func TestRuleEngine_EmptyRules(t *testing.T) {
	payload := []byte(`{"name": "test", "value": 123}`)

	tests := []struct {
		name    string
		rulestr string
		want    bool
		wantErr bool
	}{
		{
			name:    "empty string",
			rulestr: "",
			want:    true,
			wantErr: false,
		},
		{
			name:    "empty object",
			rulestr: "{}",
			want:    true,
			wantErr: false,
		},
		{
			name:    "empty rules array",
			rulestr: `{"rules":[],"groups":[]}`,
			want:    true,
			wantErr: false,
		},
		{
			name:    "empty rules list",
			rulestr: "[]",
			want:    true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RuleEngine(tt.rulestr, payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("RuleEngine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("RuleEngine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRuleEngine_EqualOperator(t *testing.T) {
	payload := []byte(`{"name": "test", "status": "active", "count": 5}`)

	tests := []struct {
		name    string
		rules   []Rule
		want    bool
		wantErr bool
	}{
		{
			name: "equal match",
			rules: []Rule{
				{ID: "1", Variable: "name", Operator: "equal_to", Value: "test"},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "equal no match",
			rules: []Rule{
				{ID: "1", Variable: "name", Operator: "equal_to", Value: "other"},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "equal with number as string",
			rules: []Rule{
				{ID: "1", Variable: "count", Operator: "equal_to", Value: "5"},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "multiple rules all match",
			rules: []Rule{
				{ID: "1", Variable: "name", Operator: "equal_to", Value: "test"},
				{ID: "2", Variable: "status", Operator: "equal_to", Value: "active"},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "multiple rules one fails",
			rules: []Rule{
				{ID: "1", Variable: "name", Operator: "equal_to", Value: "test"},
				{ID: "2", Variable: "status", Operator: "equal_to", Value: "inactive"},
			},
			want:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rulesJSON, _ := json.Marshal(tt.rules)
			got, err := RuleEngine(string(rulesJSON), payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("RuleEngine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("RuleEngine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRuleEngine_NotEqualOperator(t *testing.T) {
	payload := []byte(`{"name": "test", "status": "active"}`)

	tests := []struct {
		name    string
		rules   []Rule
		want    bool
		wantErr bool
	}{
		{
			name: "not equal match",
			rules: []Rule{
				{ID: "1", Variable: "name", Operator: "not_equal_to", Value: "other"},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "not equal no match",
			rules: []Rule{
				{ID: "1", Variable: "name", Operator: "not_equal_to", Value: "test"},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "not equal with missing field",
			rules: []Rule{
				{ID: "1", Variable: "missing", Operator: "not_equal_to", Value: "something"},
			},
			want:    true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rulesJSON, _ := json.Marshal(tt.rules)
			got, err := RuleEngine(string(rulesJSON), payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("RuleEngine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("RuleEngine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRuleEngine_ContainsOperator(t *testing.T) {
	payload := []byte(`{"message": "Hello World", "tags": ["red", "blue", "green"]}`)

	tests := []struct {
		name    string
		rules   []Rule
		want    bool
		wantErr bool
	}{
		{
			name: "contains match",
			rules: []Rule{
				{ID: "1", Variable: "message", Operator: "contains", Value: "World"},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "contains no match",
			rules: []Rule{
				{ID: "1", Variable: "message", Operator: "contains", Value: "Universe"},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "not contains match",
			rules: []Rule{
				{ID: "1", Variable: "message", Operator: "not_contains", Value: "Universe"},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "not contains no match",
			rules: []Rule{
				{ID: "1", Variable: "message", Operator: "not_contains", Value: "World"},
			},
			want:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rulesJSON, _ := json.Marshal(tt.rules)
			got, err := RuleEngine(string(rulesJSON), payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("RuleEngine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("RuleEngine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRuleEngine_NumericOperators(t *testing.T) {
	payload := []byte(`{"count": 10, "price": 99.99, "score": 50}`)

	tests := []struct {
		name    string
		rules   []Rule
		want    bool
		wantErr bool
	}{
		{
			name: "greater than match",
			rules: []Rule{
				{ID: "1", Variable: "count", Operator: "greater_than", Value: "5"},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "greater than no match",
			rules: []Rule{
				{ID: "1", Variable: "count", Operator: "greater_than", Value: "15"},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "less than match",
			rules: []Rule{
				{ID: "1", Variable: "count", Operator: "less_than", Value: "15"},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "less than no match",
			rules: []Rule{
				{ID: "1", Variable: "count", Operator: "less_than", Value: "5"},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "greater than or equal match",
			rules: []Rule{
				{ID: "1", Variable: "count", Operator: "greater_than_or_equal", Value: "10"},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "less than or equal match",
			rules: []Rule{
				{ID: "1", Variable: "count", Operator: "less_than_or_equal", Value: "10"},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "decimal comparison",
			rules: []Rule{
				{ID: "1", Variable: "price", Operator: "greater_than", Value: "50.0"},
			},
			want:    true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rulesJSON, _ := json.Marshal(tt.rules)
			got, err := RuleEngine(string(rulesJSON), payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("RuleEngine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("RuleEngine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRuleEngine_DateOperators(t *testing.T) {
	payload := []byte(`{"created": "2024-01-15T10:00:00Z", "updated": "2024-06-20T15:30:00Z"}`)

	tests := []struct {
		name    string
		rules   []Rule
		want    bool
		wantErr bool
	}{
		{
			name: "after match",
			rules: []Rule{
				{ID: "1", Variable: "created", Operator: "after", Value: "2024-01-01T00:00:00Z"},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "after no match",
			rules: []Rule{
				{ID: "1", Variable: "created", Operator: "after", Value: "2024-02-01T00:00:00Z"},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "before match",
			rules: []Rule{
				{ID: "1", Variable: "created", Operator: "before", Value: "2024-02-01T00:00:00Z"},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "before no match",
			rules: []Rule{
				{ID: "1", Variable: "created", Operator: "before", Value: "2024-01-01T00:00:00Z"},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "date format ISO8601",
			rules: []Rule{
				{ID: "1", Variable: "created", Operator: "after", Value: "2024-01-01"},
			},
			want:    true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rulesJSON, _ := json.Marshal(tt.rules)
			got, err := RuleEngine(string(rulesJSON), payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("RuleEngine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("RuleEngine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRuleEngine_LogicalGroupAND(t *testing.T) {
	payload := []byte(`{"name": "test", "status": "active", "count": 10}`)

	tests := []struct {
		name    string
		rules   []Rule
		want    bool
		wantErr bool
	}{
		{
			name: "AND group all match",
			rules: []Rule{
				{
					ID:       "group1",
					Variable: "$logical",
					Operator: "group",
					Value:    "AND",
				},
				{
					ID:       "rule1",
					Variable: "name",
					Operator: "equal_to",
					Value:    "test",
					ParentID: "group1",
				},
				{
					ID:       "rule2",
					Variable: "status",
					Operator: "equal_to",
					Value:    "active",
					ParentID: "group1",
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "AND group one fails",
			rules: []Rule{
				{
					ID:       "group1",
					Variable: "$logical",
					Operator: "group",
					Value:    "AND",
				},
				{
					ID:       "rule1",
					Variable: "name",
					Operator: "equal_to",
					Value:    "test",
					ParentID: "group1",
				},
				{
					ID:       "rule2",
					Variable: "status",
					Operator: "equal_to",
					Value:    "inactive",
					ParentID: "group1",
				},
			},
			want:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rulesJSON, _ := json.Marshal(tt.rules)
			got, err := RuleEngine(string(rulesJSON), payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("RuleEngine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("RuleEngine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRuleEngine_LogicalGroupOR(t *testing.T) {
	payload := []byte(`{"name": "test", "status": "active", "count": 10}`)

	tests := []struct {
		name    string
		rules   []Rule
		want    bool
		wantErr bool
	}{
		{
			name: "OR group one matches",
			rules: []Rule{
				{
					ID:       "group1",
					Variable: "$logical",
					Operator: "group",
					Value:    "OR",
				},
				{
					ID:       "rule1",
					Variable: "name",
					Operator: "equal_to",
					Value:    "test",
					ParentID: "group1",
				},
				{
					ID:       "rule2",
					Variable: "status",
					Operator: "equal_to",
					Value:    "inactive",
					ParentID: "group1",
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "OR group all match",
			rules: []Rule{
				{
					ID:       "group1",
					Variable: "$logical",
					Operator: "group",
					Value:    "OR",
				},
				{
					ID:       "rule1",
					Variable: "name",
					Operator: "equal_to",
					Value:    "test",
					ParentID: "group1",
				},
				{
					ID:       "rule2",
					Variable: "status",
					Operator: "equal_to",
					Value:    "active",
					ParentID: "group1",
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "OR group none match",
			rules: []Rule{
				{
					ID:       "group1",
					Variable: "$logical",
					Operator: "group",
					Value:    "OR",
				},
				{
					ID:       "rule1",
					Variable: "name",
					Operator: "equal_to",
					Value:    "other",
					ParentID: "group1",
				},
				{
					ID:       "rule2",
					Variable: "status",
					Operator: "equal_to",
					Value:    "inactive",
					ParentID: "group1",
				},
			},
			want:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rulesJSON, _ := json.Marshal(tt.rules)
			got, err := RuleEngine(string(rulesJSON), payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("RuleEngine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("RuleEngine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRuleEngine_NestedGroups(t *testing.T) {
	payload := []byte(`{"name": "test", "status": "active", "count": 10, "type": "premium"}`)

	tests := []struct {
		name    string
		rules   []Rule
		want    bool
		wantErr bool
	}{
		{
			name: "nested AND groups",
			rules: []Rule{
				{
					ID:       "group1",
					Variable: "$logical",
					Operator: "group",
					Value:    "AND",
				},
				{
					ID:       "rule1",
					Variable: "name",
					Operator: "equal_to",
					Value:    "test",
					ParentID: "group1",
				},
				{
					ID:       "group2",
					Variable: "$logical",
					Operator: "group",
					Value:    "AND",
					ParentID: "group1",
				},
				{
					ID:       "rule2",
					Variable: "status",
					Operator: "equal_to",
					Value:    "active",
					ParentID: "group2",
				},
				{
					ID:       "rule3",
					Variable: "count",
					Operator: "greater_than",
					Value:    "5",
					ParentID: "group2",
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "OR group with AND group inside",
			rules: []Rule{
				{
					ID:       "group1",
					Variable: "$logical",
					Operator: "group",
					Value:    "OR",
				},
				{
					ID:       "rule1",
					Variable: "name",
					Operator: "equal_to",
					Value:    "other",
					ParentID: "group1",
				},
				{
					ID:       "group2",
					Variable: "$logical",
					Operator: "group",
					Value:    "AND",
					ParentID: "group1",
				},
				{
					ID:       "rule2",
					Variable: "status",
					Operator: "equal_to",
					Value:    "active",
					ParentID: "group2",
				},
				{
					ID:       "rule3",
					Variable: "count",
					Operator: "greater_than",
					Value:    "5",
					ParentID: "group2",
				},
			},
			want:    true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rulesJSON, _ := json.Marshal(tt.rules)
			got, err := RuleEngine(string(rulesJSON), payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("RuleEngine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("RuleEngine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRuleEngine_MultipleRootRules(t *testing.T) {
	payload := []byte(`{"name": "test", "status": "active", "count": 10}`)

	tests := []struct {
		name    string
		rules   []Rule
		want    bool
		wantErr bool
	}{
		{
			name: "multiple root rules all match",
			rules: []Rule{
				{ID: "1", Variable: "name", Operator: "equal_to", Value: "test"},
				{ID: "2", Variable: "status", Operator: "equal_to", Value: "active"},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "multiple root rules one fails",
			rules: []Rule{
				{ID: "1", Variable: "name", Operator: "equal_to", Value: "test"},
				{ID: "2", Variable: "status", Operator: "equal_to", Value: "inactive"},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "root rule and group",
			rules: []Rule{
				{ID: "1", Variable: "name", Operator: "equal_to", Value: "test"},
				{
					ID:       "group1",
					Variable: "$logical",
					Operator: "group",
					Value:    "AND",
				},
				{
					ID:       "rule2",
					Variable: "status",
					Operator: "equal_to",
					Value:    "active",
					ParentID: "group1",
				},
			},
			want:    true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rulesJSON, _ := json.Marshal(tt.rules)
			got, err := RuleEngine(string(rulesJSON), payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("RuleEngine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("RuleEngine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRuleEngine_MissingFields(t *testing.T) {
	payload := []byte(`{"name": "test"}`)

	tests := []struct {
		name    string
		rules   []Rule
		want    bool
		wantErr bool
	}{
		{
			name: "missing field with equal_to",
			rules: []Rule{
				{ID: "1", Variable: "missing", Operator: "equal_to", Value: ""},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "missing field with not_equal_to",
			rules: []Rule{
				{ID: "1", Variable: "missing", Operator: "not_equal_to", Value: "something"},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "missing field with contains",
			rules: []Rule{
				{ID: "1", Variable: "missing", Operator: "contains", Value: "test"},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "missing field with greater_than",
			rules: []Rule{
				{ID: "1", Variable: "missing", Operator: "greater_than", Value: "5"},
			},
			want:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rulesJSON, _ := json.Marshal(tt.rules)
			got, err := RuleEngine(string(rulesJSON), payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("RuleEngine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("RuleEngine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRuleEngine_ComplexNestedScenario(t *testing.T) {
	payload := []byte(`{
		"user": {
			"name": "John",
			"age": 30,
			"email": "john@example.com"
		},
		"order": {
			"total": 150.50,
			"status": "completed"
		},
		"timestamp": "2024-06-15T10:00:00Z"
	}`)

	tests := []struct {
		name    string
		rules   []Rule
		want    bool
		wantErr bool
	}{
		{
			name: "complex nested path evaluation",
			rules: []Rule{
				{
					ID:       "group1",
					Variable: "$logical",
					Operator: "group",
					Value:    "AND",
				},
				{
					ID:       "rule1",
					Variable: "user.name",
					Operator: "equal_to",
					Value:    "John",
					ParentID: "group1",
				},
				{
					ID:       "rule2",
					Variable: "user.age",
					Operator: "greater_than",
					Value:    "18",
					ParentID: "group1",
				},
				{
					ID:       "rule3",
					Variable: "order.total",
					Operator: "greater_than",
					Value:    "100",
					ParentID: "group1",
				},
				{
					ID:       "rule4",
					Variable: "order.status",
					Operator: "equal_to",
					Value:    "completed",
					ParentID: "group1",
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "OR group with nested paths",
			rules: []Rule{
				{
					ID:       "group1",
					Variable: "$logical",
					Operator: "group",
					Value:    "OR",
				},
				{
					ID:       "rule1",
					Variable: "user.name",
					Operator: "equal_to",
					Value:    "Jane",
					ParentID: "group1",
				},
				{
					ID:       "rule2",
					Variable: "order.status",
					Operator: "equal_to",
					Value:    "completed",
					ParentID: "group1",
				},
			},
			want:    true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rulesJSON, _ := json.Marshal(tt.rules)
			got, err := RuleEngine(string(rulesJSON), payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("RuleEngine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("RuleEngine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRuleEngine_InvalidJSON(t *testing.T) {
	payload := []byte(`{"name": "test"}`)

	tests := []struct {
		name    string
		rulestr string
		want    bool
		wantErr bool
	}{
		{
			name:    "invalid JSON",
			rulestr: `{invalid json}`,
			want:    false,
			wantErr: true,
		},
		{
			name:    "malformed rules",
			rulestr: `[{"id": "1", "variable": "name"}]`, // missing operator and value
			want:    false,
			wantErr: false, // Should not error, but may fail evaluation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RuleEngine(tt.rulestr, payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("RuleEngine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("RuleEngine() = %v, want %v", got, tt.want)
			}
		})
	}
}

