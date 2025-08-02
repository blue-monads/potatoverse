package buddy

import (
	"strings"

	"github.com/k0kubun/pp"
	"github.com/spaolacci/murmur3"
	"github.com/upper/db/v4"
)

type ChangeCapureService struct {
	db db.Session
}

type Schema struct {
	Type     string  `json:"type" db:"type,omitempty"`
	Name     string  `json:"name" db:"name"`
	TblName  string  `json:"tbl_name" db:"tbl_name"`
	RootPage int64   `json:"rootpage" db:"rootpage"`
	Sql      *string `json:"sql" db:"sql"`
}

func New(sess db.Session) *ChangeCapureService {
	return &ChangeCapureService{
		db: sess,
	}
}

func (c *ChangeCapureService) CaptureHooks() error {

	schemas := make([]Schema, 0)

	err := c.db.Collection("sqlite_master").Find().All(&schemas)
	if err != nil {
		return err
	}

	// pp.Println("@schema", schemas)

	//	triggersIndex := make(map[string]string)

	hash := murmur3.New32()
	hash.Write([]byte(`hashashhsa`))

	pp.Println("@", convert32bitToString(hash.Sum(nil)))

	return nil

}

var charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func convert32bitToString(data []byte) string {
	charsetLength := len(charset)
	var result strings.Builder

	for _, b := range data {
		highBits := int(b >> 2)
		lowBits := int(b & 0x03)
		highIndex := highBits % charsetLength
		lowIndex := lowBits % charsetLength
		result.WriteByte(charset[highIndex])
		if lowBits > 0 {
			result.WriteByte(charset[lowIndex])
		}
	}

	return result.String()
}
