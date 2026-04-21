package core

// DB Operations
type DBQueryReq struct {
	Query string `json:"query"`
	Args  []any  `json:"args"`
}

type DBInsertReq struct {
	Table string         `json:"table"`
	Data  map[string]any `json:"data"`
}

type DBUpdateByIdReq struct {
	Table string         `json:"table"`
	ID    int64          `json:"id"`
	Data  map[string]any `json:"data"`
}

type DBIdReq struct {
	Table string `json:"table"`
	ID    int64  `json:"id"`
}

type DBUpdateByCondReq struct {
	Table string         `json:"table"`
	Cond  map[string]any `json:"cond"`
	Data  map[string]any `json:"data"`
}

type DBCondReq struct {
	Table string         `json:"table"`
	Cond  map[string]any `json:"cond"`
}

// KV Operations
type KVQueryReq struct {
	Cond         map[string]any `json:"cond"`
	Offset       int            `json:"offset"`
	Limit        int            `json:"limit"`
	IncludeValue bool           `json:"include_value"`
}

type KVKeyReq struct {
	Group string `json:"group"`
	Key   string `json:"key"`
}

type KVDataReq struct {
	Group string         `json:"group"`
	Key   string         `json:"key"`
	Data  map[string]any `json:"data"`
}

// Core Operations
type PublishEventOptions struct {
	Name        string `json:"name"`
	Payload     any    `json:"payload"`
	ResourceId  string `json:"resource_id"`
	CollapseKey string `json:"collapse_key"`
}

type SignFsPresignedTokenOptions struct {
	Path     string `json:"path"`
	FileName string `json:"file_name"`
	UserId   int64  `json:"user_id"`
}

type SignAdviseryTokenOptions struct {
	TokenSubType string         `json:"token_sub_type"`
	UserId       int64          `json:"user_id"`
	Data         map[string]any `json:"data"`
}

type ParseTokenReq struct {
	Token string `json:"token"`
}

// Capability Operations
type CapTokenSignOptions struct {
	ResourceId string         `json:"resource_id"`
	ExtraMeta  map[string]any `json:"extrameta"`
	UserId     int64          `json:"user_id"`
	SubType    string         `json:"sub_type"`
}
