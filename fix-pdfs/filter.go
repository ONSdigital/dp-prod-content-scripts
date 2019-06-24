package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ONSdigital/dp-zebedee-utils/collections"
	"github.com/ONSdigital/dp-zebedee-utils/content"
	"github.com/ONSdigital/log.go/log"
)

type fix struct {
	OutputPath  string
	MasterDir   string
	AllCols     *collections.Collections
	FixC        *collections.Collection
	Limit       int
	Blocked     []string
	PageTypes   map[string]bool
	AnyPageType bool
	FixCount    int
	CSVW        *csv.Writer
}

func (f *fix) Filter(path string, info os.FileInfo) (bool, error) {
	if info.IsDir() {
		return false, nil
	}

	if strings.Contains(path, "/previous/") {
		return false, nil
	}

	if info.Name() != "data.json" && info.Name() != "data_cy.json" {
		return false, nil
	}

	jBytes, err := content.ReadJson(path)
	if err != nil {
		return false, err
	}

	if !strings.Contains(string(jBytes), oldEmail) {
		return false, nil
	}

	// if empty then consider all page types.
	if len(f.PageTypes) == 0 {
		return true, nil
	}

	// else check the type
	pageType, err := content.GetPageType(jBytes)
	if err != nil {
		return false, err
	}

	if _, ok := f.PageTypes[pageType.Value]; !ok {
		return false, nil
	}

	return true, nil
}

func (f *fix) Process(path string) error {
	jBytes, err := content.ReadJson(path)
	if err != nil {
		return err
	}

	jsonStr := string(jBytes)

	uri, err := filepath.Rel(f.MasterDir, path)
	if err != nil {
		return err
	}

	uri = "/" + uri
	if blocked, name := f.AllCols.IsBlocked(uri); blocked {
		f.Blocked = append(f.Blocked, fmt.Sprintf("%s:%s", name, uri))
		return nil
	}

	jsonStr = strings.Replace(jsonStr, oldEmail, newEmail, -1)

	if err := f.FixC.AddToReviewed(uri, []byte(jsonStr)); err != nil {
		return err
	}

	pageType, err := content.GetPageType([]byte(jsonStr))
	if err != nil {
		return err
	}

	if err := f.logFix(uri, pageType); err != nil {
		return err
	}

	return nil
}

func (f *fix) OnComplete() error {
	f.CSVW.Flush()

	total, stats, err := f.getResults()
	if err != nil {
		return err
	}

	logD := log.Data{
		"stats":          stats,
		"fix_count":      total,
		"fix_collection": f.FixC.Name,
		"blocked":        f.Blocked,
	}

	logD["blocked"] = len(f.Blocked)

	log.Event(nil, "script fixing content completed successfully", logD)
	return nil
}

func (f *fix) getResults() (int, map[string]int, error) {
	file, err := os.Open(f.OutputPath)
	if err != nil {
		return 0, nil, err
	}
	defer file.Close()

	csvR := csv.NewReader(file)
	rec, err := csvR.ReadAll()
	if err != nil {
		return 0, nil, err
	}

	stats := make(map[string]int, 0)
	for i, row := range rec {
		if i == 0 {
			continue
		}

		if count, ok := stats[row[1]]; ok {
			stats[row[1]] = count + 1
		} else {
			stats[row[1]] = 1
		}
	}

	return len(rec) - 1, stats, nil
}

func (f *fix) LimitReached() bool {
	if f.Limit == -1 {
		return false
	}
	return f.FixCount >= f.Limit
}

func (f *fix) logFix(uri string, pageType *content.PageType) error {
	f.FixCount += 1
	err := f.CSVW.Write(toCSVRow(uri, pageType))
	return err
}

func toCSVRow(uri string, pageType *content.PageType) []string {
	return []string{uri, pageType.Value, strconv.FormatBool(isPDF(pageType))}
}

func isPDF(pageType *content.PageType) bool {
	switch pageType.Value {
	case "article", "bulletin", "compendium_landing_page", "compendium_chapter", "static_methodology":
		return true
	}
	return false
}
