package ccurd

import "regexp"

type Methods struct {
	Table      string                     `json:"table"`
	Mode       string                     `json:"mode"`
	EventName  string                     `json:"event_name"`
	Validators map[string]*ValidationItem `json:"validators"`
}

type ValidationItem struct {
	Type         string `json:"type"`
	Required     bool   `json:"required"`
	MinLength    int64  `json:"min_length"`
	MaxLength    int64  `json:"max_length"`
	RegexPattern string `json:"regex"`

	compiledRegex *regexp.Regexp `json:"-"`
}

/*

   "stmt": "insert into articles (title, content) values (?, ?)",
   "allowed_fields": ["id", "title", "content"],
   "mode": "insert",
   "validator": {
       "title": {
           "type": "string",
           "required": true,
           "min_length": 3,
           "max_length": 100
       },
       "content": {
           "type": "string",
           "required": true,
           "regex": "^[a-zA-Z0-9]+$"
       },
   }


*/
