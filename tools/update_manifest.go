package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/blue-monads/potatoverse/backend/xtypes/models"
	"github.com/pelletier/go-toml/v2"
)

func main() {
	// Find all potato.toml and potato.json files in potatoverse and potato-apps
	var manifestFiles []string

	// Search in potatoverse (current directory is tools, go up one level)
	err := filepath.Walk("..", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && (info.Name() == "potato-apps" || strings.Contains(path, ".git")) {
			return filepath.SkipDir // Skip potato-apps (we'll search it separately) and git directories
		}
		if !info.IsDir() && (strings.EqualFold(info.Name(), "potato.toml") || strings.EqualFold(info.Name(), "potato.json")) {
			manifestFiles = append(manifestFiles, path)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Error walking potatoverse directory: %v\n", err)
	}

	// Search in potato-apps (go up one level and into potato-apps)
	err = filepath.Walk("../potato-apps", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && strings.Contains(path, ".git") {
			return filepath.SkipDir // Skip git directories
		}
		if !info.IsDir() && (strings.EqualFold(info.Name(), "potato.toml") || strings.EqualFold(info.Name(), "potato.json")) {
			manifestFiles = append(manifestFiles, path)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Error walking potato-apps directory: %v\n", err)
	}

	// Process each manifest file
	for _, filePath := range manifestFiles {
		fmt.Printf("Processing: %s\n", filePath)

		if strings.HasSuffix(strings.ToLower(filePath), ".json") {
			err = processJSONFile(filePath)
		} else if strings.HasSuffix(strings.ToLower(filePath), ".toml") {
			err = processTOMLFile(filePath)
		}

		if err != nil {
			fmt.Printf("Error processing %s: %v\n", filePath, err)
		} else {
			fmt.Printf("Successfully updated: %s\n", filePath)
		}
	}

	fmt.Println("\nAll files processed.")
}

func processJSONFile(filePath string) error {
	// Read and parse existing JSON file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var oldData map[string]any
	err = json.Unmarshal(data, &oldData)
	if err != nil {
		return err
	}

	// Create new V2 structure
	pkg := &models.PotatoPackage{
		Name:          getStringValue(oldData, "name"),
		Slug:          getStringValue(oldData, "slug"),
		Info:          getStringValue(oldData, "info"),
		CanonicalUrl:  getStringValue(oldData, "canonical_url"),
		Tags:          getStringSlice(oldData, "tags"),
		FormatVersion: "0.0.2", // Update format version
		AuthorName:    getStringValue(oldData, "author_name"),
		AuthorEmail:   getStringValue(oldData, "author_email"),
		AuthorSite:    getStringValue(oldData, "author_site"),
		SourceCode:    getStringValue(oldData, "source_code"),
		License:       getStringValue(oldData, "license"),
		Version:       getStringValue(oldData, "version"),
		SpecialPages:  make(map[string]string),
		Spaces:        []models.PotatoSpace{},
		Capabilities:  []models.PotatoCapability{},
	}

	// Handle special pages
	if initPage := getStringValue(oldData, "init_page"); initPage != "" {
		pkg.SpecialPages["init_page"] = initPage
	}
	if updatePage := getStringValue(oldData, "update_page"); updatePage != "" {
		pkg.SpecialPages["update_page"] = updatePage
	}

	// Handle artifacts (convert to spaces and capabilities)
	if artifacts, ok := oldData["artifacts"].([]interface{}); ok {
		for _, artifact := range artifacts {
			if artMap, ok := artifact.(map[string]interface{}); ok {
				kind := getStringValue(artMap, "kind")
				if kind == "space" {
					space := models.PotatoSpace{
						Namespace:       getStringValue(artMap, "namespace"),
						ExecutorType:    getStringValue(artMap, "executor_type"),
						ExecutorSubType: getStringValue(artMap, "executor_sub_type"),
						ServerFile:      getStringValue(artMap, "server_file"),
						DevServePort:    getIntValue(artMap, "dev_serve_port"),
						IsDefault:       pkg.Slug == getStringValue(artMap, "namespace"),
					}

					// Handle route options
					if routeOpts, ok := artMap["route_options"].(map[string]interface{}); ok {
						space.RouteOptions = models.PotatoRouteOptions{
							RouterType:         getStringValue(routeOpts, "router_type"),
							ServeFolder:        getStringValue(routeOpts, "serve_folder"),
							ForceHtmlExtension: getBoolValue(routeOpts, "force_html_extension"),
							ForceIndexHtmlFile: getBoolValue(routeOpts, "force_index_html_file"),
							OnNotFoundFile:     getStringValue(routeOpts, "on_not_found_file"),
							TrimPathPrefix:     getStringValue(routeOpts, "trim_path_prefix"),
							TemplateFolder:     getStringValue(routeOpts, "template_folder"),
						}

						if routes, ok := routeOpts["routes"].([]interface{}); ok {
							for _, r := range routes {
								if routeMap, ok := r.(map[string]interface{}); ok {
									space.RouteOptions.Routes = append(space.RouteOptions.Routes, models.PotatoRoute{
										Path:    getStringValue(routeMap, "path"),
										Method:  getStringValue(routeMap, "method"),
										Type:    getStringValue(routeMap, "type"),
										Handler: getStringValue(routeMap, "handler"),
										File:    getStringValue(routeMap, "file"),
									})
								}
							}
						}
					}

					pkg.Spaces = append(pkg.Spaces, space)
				} else if kind == "capability" {
					capability := models.PotatoCapability{
						Name:    getStringValue(artMap, "name"),
						Type:    getStringValue(artMap, "type"),
						Options: make(map[string]interface{}),
						Spaces:  getStringSlice(artMap, "spaces"),
					}

					if options, ok := artMap["options"].(map[string]interface{}); ok {
						capability.Options = options
					}

					pkg.Capabilities = append(pkg.Capabilities, capability)
				}
			}
		}
	}

	// Handle developer options
	if devOptions, ok := oldData["developer"].(map[string]interface{}); ok {
		pkg.Developer = &models.DeveloperOptions{
			ServerUrl:     getStringValue(devOptions, "server_url"),
			Token:         getStringValue(devOptions, "token"),
			TokenEnv:      getStringValue(devOptions, "token_env"),
			OutputZipFile: getStringValue(devOptions, "output_zip_file"),
			IncludeFiles:  getStringSlice(devOptions, "include_files"),
			ExcludeFiles:  getStringSlice(devOptions, "exclude_files"),
			BuildCommand:  getStringValue(devOptions, "build_command"),
		}
	} else if devOptions, ok := oldData["dev_options"].(map[string]interface{}); ok {
		// Handle old dev_options field name
		pkg.Developer = &models.DeveloperOptions{
			ServerUrl: getStringValue(devOptions, "server_url"),
			Token:     getStringValue(devOptions, "token"),
			TokenEnv:  getStringValue(devOptions, "token_env"),
		}
	}

	// Write updated file
	var newData []byte
	if strings.HasSuffix(strings.ToLower(filePath), ".json") {
		newData, err = json.MarshalIndent(pkg, "", "    ")
	} else {
		newData, err = toml.Marshal(pkg)
	}
	if err != nil {
		return err
	}

	err = os.WriteFile(filePath, newData, 0644)
	if err != nil {
		return err
	}

	return nil
}

func processTOMLFile(filePath string) error {
	// Read and parse existing TOML file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var oldData map[string]any
	err = toml.Unmarshal(data, &oldData)
	if err != nil {
		return err
	}

	// Create new V2 structure
	pkg := &models.PotatoPackage{
		Name:          getStringValue(oldData, "name"),
		Slug:          getStringValue(oldData, "slug"),
		Info:          getStringValue(oldData, "info"),
		CanonicalUrl:  getStringValue(oldData, "canonical_url"),
		Tags:          getStringSlice(oldData, "tags"),
		FormatVersion: "0.0.2", // Update format version
		AuthorName:    getStringValue(oldData, "author_name"),
		AuthorEmail:   getStringValue(oldData, "author_email"),
		AuthorSite:    getStringValue(oldData, "author_site"),
		SourceCode:    getStringValue(oldData, "source_code"),
		License:       getStringValue(oldData, "license"),
		Version:       getStringValue(oldData, "version"),
		SpecialPages:  make(map[string]string),
		Spaces:        []models.PotatoSpace{},
		Capabilities:  []models.PotatoCapability{},
	}

	// Handle special pages
	if initPage := getStringValue(oldData, "init_page"); initPage != "" {
		pkg.SpecialPages["init_page"] = initPage
	}
	if updatePage := getStringValue(oldData, "update_page"); updatePage != "" {
		pkg.SpecialPages["update_page"] = updatePage
	}

	// Handle artifacts (convert to spaces and capabilities)
	if artifacts, ok := oldData["artifacts"].([]interface{}); ok {
		for _, artifact := range artifacts {
			if artMap, ok := artifact.(map[string]interface{}); ok {
				kind := getStringValue(artMap, "kind")
				if kind == "space" {
					space := models.PotatoSpace{
						Namespace:       getStringValue(artMap, "namespace"),
						ExecutorType:    getStringValue(artMap, "executor_type"),
						ExecutorSubType: getStringValue(artMap, "executor_sub_type"),
						ServerFile:      getStringValue(artMap, "server_file"),
						DevServePort:    getIntValue(artMap, "dev_serve_port"),
						IsDefault:       pkg.Slug == getStringValue(artMap, "namespace"),
					}

					// Handle route options
					if routeOpts, ok := artMap["route_options"].(map[string]interface{}); ok {
						space.RouteOptions = models.PotatoRouteOptions{
							RouterType:         getStringValue(routeOpts, "router_type"),
							ServeFolder:        getStringValue(routeOpts, "serve_folder"),
							ForceHtmlExtension: getBoolValue(routeOpts, "force_html_extension"),
							ForceIndexHtmlFile: getBoolValue(routeOpts, "force_index_html_file"),
							OnNotFoundFile:     getStringValue(routeOpts, "on_not_found_file"),
							TrimPathPrefix:     getStringValue(routeOpts, "trim_path_prefix"),
							TemplateFolder:     getStringValue(routeOpts, "template_folder"),
						}

						if routes, ok := routeOpts["routes"].([]interface{}); ok {
							for _, r := range routes {
								if routeMap, ok := r.(map[string]interface{}); ok {
									space.RouteOptions.Routes = append(space.RouteOptions.Routes, models.PotatoRoute{
										Path:    getStringValue(routeMap, "path"),
										Method:  getStringValue(routeMap, "method"),
										Type:    getStringValue(routeMap, "type"),
										Handler: getStringValue(routeMap, "handler"),
										File:    getStringValue(routeMap, "file"),
									})
								}
							}
						}
					}

					pkg.Spaces = append(pkg.Spaces, space)
				} else if kind == "capability" {
					capability := models.PotatoCapability{
						Name:    getStringValue(artMap, "name"),
						Type:    getStringValue(artMap, "type"),
						Options: make(map[string]interface{}),
						Spaces:  getStringSlice(artMap, "spaces"),
					}

					if options, ok := artMap["options"].(map[string]interface{}); ok {
						capability.Options = options
					}

					pkg.Capabilities = append(pkg.Capabilities, capability)
				}
			}
		}
	}

	// Handle developer options
	if devOptions, ok := oldData["developer"].(map[string]interface{}); ok {
		pkg.Developer = &models.DeveloperOptions{
			ServerUrl:     getStringValue(devOptions, "server_url"),
			Token:         getStringValue(devOptions, "token"),
			TokenEnv:      getStringValue(devOptions, "token_env"),
			OutputZipFile: getStringValue(devOptions, "output_zip_file"),
			IncludeFiles:  getStringSlice(devOptions, "include_files"),
			ExcludeFiles:  getStringSlice(devOptions, "exclude_files"),
			BuildCommand:  getStringValue(devOptions, "build_command"),
		}
	} else if devOptions, ok := oldData["dev_options"].(map[string]interface{}); ok {
		// Handle old dev_options field name
		pkg.Developer = &models.DeveloperOptions{
			ServerUrl: getStringValue(devOptions, "server_url"),
			Token:     getStringValue(devOptions, "token"),
			TokenEnv:  getStringValue(devOptions, "token_env"),
		}
	} else if packagingOptions, ok := oldData["packaging"].(map[string]interface{}); ok {
		// Handle old packaging field
		if pkg.Developer == nil {
			pkg.Developer = &models.DeveloperOptions{}
		}
		if outputZip, ok := packagingOptions["output_zip_file"].(string); ok {
			pkg.Developer.OutputZipFile = outputZip
		}
		if includeFiles, ok := packagingOptions["include_files"].([]interface{}); ok {
			for _, f := range includeFiles {
				if s, ok := f.(string); ok {
					pkg.Developer.IncludeFiles = append(pkg.Developer.IncludeFiles, s)
				}
			}
		}
		if excludeFiles, ok := packagingOptions["exclude_files"].([]interface{}); ok {
			for _, f := range excludeFiles {
				if s, ok := f.(string); ok {
					pkg.Developer.ExcludeFiles = append(pkg.Developer.ExcludeFiles, s)
				}
			}
		}
	}

	// Write updated file
	var newData []byte
	if strings.HasSuffix(strings.ToLower(filePath), ".json") {
		newData, err = json.MarshalIndent(pkg, "", "    ")
	} else {
		newData, err = toml.Marshal(pkg)
	}
	if err != nil {
		return err
	}

	err = os.WriteFile(filePath, newData, 0644)
	if err != nil {
		return err
	}

	return nil
}

func getStringValue(data map[string]any, key string) string {
	if value, ok := data[key]; ok {
		if s, ok := value.(string); ok {
			return s
		}
	}
	return ""
}

func getIntValue(data map[string]any, key string) int {
	if value, ok := data[key]; ok {
		switch v := value.(type) {
		case int:
			return v
		case float64:
			return int(v)
		}
	}
	return 0
}

func getBoolValue(data map[string]any, key string) bool {
	if value, ok := data[key]; ok {
		if b, ok := value.(bool); ok {
			return b
		}
	}
	return false
}

func getStringSlice(data map[string]any, key string) []string {
	var slice []string
	if value, ok := data[key]; ok {
		if arr, ok := value.([]interface{}); ok {
			for _, item := range arr {
				if s, ok := item.(string); ok {
					slice = append(slice, s)
				}
			}
		}
	}
	return slice
}
