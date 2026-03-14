package corehub

import (
	"archive/zip"
	"bufio"
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

const exportBatchSize = 10

func (c *CoreHub) ExportState(opts *StateExport) (string, error) {

	zfile, err := os.CreateTemp("", "export_state_*.zip")
	if err != nil {
		return "", err
	}

	defer zfile.Close()

	dataOpts := c.db.GetLowPackageDBOps(opts.InstallId)
	tableInfos, err := dataOpts.ListTables()
	if err != nil {
		return "", err
	}

	qq.Println("@tables", tableInfos)

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

			qq.Println("@exportTableData", table, maxRowId, exportBatchSize)

			data, err := dataOpts.FindAllByQuery(&datahub.FindQuery{
				Table: table,
				Limit: exportBatchSize,
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

	qq.Println("@tables", tableInfos)

	for _, tableInfo := range tableInfos {

		if tableInfo.TableType == "virtual" {
			continue
		}

		if tableInfo.TableType == "virtual_sub_type" {
			continue
		}

		table := tableInfo.Name

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

func (c *CoreHub) Import(installId int64, zipfile string) error {

	zfile, err := zip.OpenReader(zipfile)
	if err != nil {
		return err
	}
	defer zfile.Close()

	dataOpts := c.db.GetLowPackageDBOps(installId)

	for _, file := range zfile.File {

		if file.FileInfo().Size() == 0 {
			continue
		}

		if !strings.HasSuffix(file.Name, ".jsonl") {
			continue
		}

		tableName := strings.TrimSuffix(file.Name, ".jsonl")
		qq.Println("@importTable", tableName)

		reader, err := file.Open()
		if err != nil {
			return err
		}
		defer reader.Close()

		// Read and insert each line
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			line := scanner.Bytes()
			if len(line) == 0 {
				continue
			}

			var rowData map[string]any
			err := json.Unmarshal(line, &rowData)
			if err != nil {
				return err
			}

			// Insert the row into the table
			_, err = dataOpts.Insert(tableName, rowData)
			if err != nil {
				return err
			}
		}

		if err := scanner.Err(); err != nil {
			return err
		}
	}

	return nil
}
