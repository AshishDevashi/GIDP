// Command export_postman dumps all routes registered on the Gin engine into a
// Postman Collection v2.1 JSON file. Run with:
//
//	go run tools/export_postman.go
//
// Output is written to postman/collections/GIDP.postman_collection.json
package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/AshishDevashi/GIDP/internal/config"
	"github.com/AshishDevashi/GIDP/internal/platform/logger"
	"github.com/AshishDevashi/GIDP/internal/server"
)

type postmanURL struct {
	Raw  string   `json:"raw"`
	Host []string `json:"host"`
	Path []string `json:"path"`
}

type postmanRequest struct {
	Method string     `json:"method"`
	URL    postmanURL `json:"url"`
}

type postmanItem struct {
	Name    string          `json:"name"`
	Item    []*postmanItem  `json:"item,omitempty"`
	Request *postmanRequest `json:"request,omitempty"`
}

type postmanInfo struct {
	Name   string `json:"name"`
	Schema string `json:"schema"`
}

type postmanVariable struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type postmanCollection struct {
	Info     postmanInfo       `json:"info"`
	Item     []*postmanItem    `json:"item"`
	Variable []postmanVariable `json:"variable"`
}

// folderKey returns the module name a route belongs to, e.g. "/api/v1/teams/:id" -> "teams".
func folderKey(path string) string {
	segments := strings.Split(strings.Trim(path, "/"), "/")
	if len(segments) >= 3 && segments[0] == "api" && segments[1] == "v1" {
		return segments[2]
	}
	return segments[0]
}

func folderName(key string) string {
	if key == "" {
		return key
	}
	return strings.ToUpper(key[:1]) + key[1:]
}

func main() {
	cfg := config.Load()
	log := logger.New(cfg.Env)

	// db is never used at request time when we only enumerate routes.
	srv := server.New(cfg, log, nil)
	routes := srv.Router().Routes()

	sort.Slice(routes, func(i, j int) bool {
		if routes[i].Path != routes[j].Path {
			return routes[i].Path < routes[j].Path
		}
		return routes[i].Method < routes[j].Method
	})

	items := make([]*postmanItem, 0)
	folders := make(map[string]*postmanItem)
	var folderOrder []string

	for _, r := range routes {
		segments := strings.Split(strings.Trim(r.Path, "/"), "/")
		reqItem := &postmanItem{
			Name: r.Method + " " + r.Path,
			Request: &postmanRequest{
				Method: r.Method,
				URL: postmanURL{
					Raw:  "{{base_url}}" + r.Path,
					Host: []string{"{{base_url}}"},
					Path: segments,
				},
			},
		}

		key := folderKey(r.Path)
		folder, ok := folders[key]
		if !ok {
			folder = &postmanItem{Name: folderName(key)}
			folders[key] = folder
			folderOrder = append(folderOrder, key)
		}
		folder.Item = append(folder.Item, reqItem)
	}

	for _, key := range folderOrder {
		items = append(items, folders[key])
	}

	collection := postmanCollection{
		Info: postmanInfo{
			Name:   "GIDP",
			Schema: "https://schema.getpostman.com/json/collection/v2.1.0/collection.json",
		},
		Item:     items,
		Variable: []postmanVariable{{Key: "base_url", Value: "http://localhost:8080"}},
	}

	outPath := filepath.Join("postman", "collections", "GIDP.postman_collection.json")
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		log.Error("failed to create output directory", "error", err)
		os.Exit(1)
	}

	data, err := json.MarshalIndent(collection, "", "  ")
	if err != nil {
		log.Error("failed to marshal collection", "error", err)
		os.Exit(1)
	}

	if err := os.WriteFile(outPath, data, 0o644); err != nil {
		log.Error("failed to write collection file", "error", err)
		os.Exit(1)
	}

	log.Info("postman collection exported", "path", outPath, "routes", len(routes), "folders", len(items))
}
