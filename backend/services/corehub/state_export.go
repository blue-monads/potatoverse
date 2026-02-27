package corehub

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/blue-monads/potatoverse/backend/services/datahub"
	"github.com/blue-monads/potatoverse/backend/utils/qq"
)

type StateExport struct {
	InstallId     int64    `json:"install_id"`
	ExcludeTables []string `json:"exclude_tables"`
}

func (c *CoreHub) ExportState(opts *StateExport) (string, error) {

	zfile, err := os.CreateTemp("", "export_state_*.zip")
	if err != nil {
		return "", err
	}

	defer zfile.Close()

	dataOpts := c.db.GetLowPackageDBOps(opts.InstallId)
	tables, err := dataOpts.ListTables()
	if err != nil {
		return "", err
	}

	zipWriter := zip.NewWriter(zfile)

	exportTableData := func(table string) error {

		subzFile, err := zipWriter.Create(table + ".jsonl")
		if err != nil {
			return err
		}

		qq.Println("@exportTableData", table)

		maxRowId := 0

		pkColumn, err := dataOpts.FindTablePK(table)
		if err != nil {
			return err
		}

		for {

			data, err := dataOpts.FindAllByQuery(&datahub.FindQuery{
				Table: table,
				Limit: 10,
				Order: pkColumn,
				Cond: map[any]any{
					fmt.Sprintf("%s >", pkColumn): maxRowId,
				},
			})
			if err != nil {
				return err
			}
			if len(data) == 0 {
				break
			}

			for _, row := range data {
				line, err := json.Marshal(row)
				if err != nil {
					return err
				}
				_, err = subzFile.Write(line)
				if err != nil {
					return err
				}
				_, err = subzFile.Write([]byte("\n"))
				if err != nil {
					return err
				}

				// update maxRowId for the next query
				if rid, ok := row[pkColumn]; ok {
					switch v := rid.(type) {
					case int:
						if v > maxRowId {
							maxRowId = v
						}
					case int64:
						if int(v) > maxRowId {
							maxRowId = int(v)
						}
					case float64:
						if int(v) > maxRowId {
							maxRowId = int(v)
						}
					}

				} else {
					panic("Could not extract pk")
				}
			}
		}

		return nil

	}

	qq.Println("@tables", tables)

	for _, table := range tables {

		prefix := fmt.Sprintf("zz_P__%d__", opts.InstallId)
		umangledTable := table
		if after, ok := strings.CutPrefix(table, prefix); ok {
			umangledTable = after
		}

		if slices.Contains(opts.ExcludeTables, umangledTable) {
			continue
		}

		err := exportTableData(umangledTable)
		if err != nil {
			zipWriter.Close()
			return "", err
		}

	}

	// close zip writer before returning
	if err := zipWriter.Close(); err != nil {
		return "", err
	}

	return zfile.Name(), nil

}
