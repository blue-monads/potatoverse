package sqlitecore

import (
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

//go:embed migrations/*
var MigrationBox embed.FS

func Get() string {
	entries, err := fs.ReadDir(MigrationBox, "migrations")
	if err != nil {
		panic(fmt.Errorf("read migrations dir: %w", err))
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	var base strings.Builder
	for _, entry := range entries {
		data, err := MigrationBox.ReadFile("migrations/" + entry.Name())
		if err != nil {
			panic(fmt.Errorf("read migration %s: %w", entry.Name(), err))
		}
		base.Write(data)
		base.WriteString("\n")
	}

	patchedData, err := MigrationBox.ReadFile("migrations/0004_file.sql")
	if err != nil {
		panic(fmt.Errorf("read 0004_file.sql: %w", err))
	}
	patched := strings.ReplaceAll(string(patchedData), "FileMeta", "PFileMeta")
	patched = strings.ReplaceAll(patched, "FileBlob", "PFileBlob")

	return base.String() + "\n" + patched
}

func GetFileSchema() string {
	data, err := MigrationBox.ReadFile("migrations/0004_file.sql")
	if err != nil {
		panic(fmt.Errorf("read 0004_file.sql: %w", err))
	}

	return string(data)

}

func GetSigSchema() string {
	data, err := MigrationBox.ReadFile("migrations/0002_sig.sql")
	if err != nil {
		panic(fmt.Errorf("read 0002_sig.sql: %w", err))
	}

	return string(data)
}
