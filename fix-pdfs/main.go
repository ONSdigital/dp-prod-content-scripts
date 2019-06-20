package main

import (
	"errors"
	"flag"
	"os"
	"path/filepath"
	"strings"

	"github.com/ONSdigital/dp-zebedee-utils/collections"
	"github.com/ONSdigital/dp-zebedee-utils/content"
	"github.com/ONSdigital/log.go/log"
)

var (
	oldEmail = "@ons.gsi.gov.uk"
	newEmail = "@ons.gov.uk"
)

func main() {
	baseDir, collectionName, pageTypes, limit := getFlags()

	if !content.Exists(baseDir) {
		errExit(errors.New("master dir does not exist"))
	}

	collectionsDir := filepath.Join(baseDir, "collections")
	masterDir := filepath.Join(baseDir, "master")

	if collectionName == "" {
		errExit(errors.New("no collection name provided"))
	}

	fixC := collections.New(collectionsDir, collectionName)
	if err := collections.Save(fixC); err != nil {
		errExit(err)
	}

	allCols, err := collections.GetCollections(collectionsDir)
	if err != nil {
		errExit(err)
	}

	types, err := parsePageTypes(pageTypes)
	if err != nil {
		errExit(err)
	}

	job := &fix{
		Limit:     limit,
		FixCount:  0,
		FixLog:    make(map[string]int, 0),
		MasterDir: masterDir,
		AllCols:   allCols,
		FixC:      fixC,
		Blocked:   make([]string, 0),
		PageTypes: types,
	}

	if err = content.FilterAndProcess(masterDir, job); err != nil {
		errExit(err)
	}
}

func getFlags() (string, string, string, int) {
	baseDir := flag.String("dir", "", "the zebedee master dir")
	collectionName := flag.String("col", "", "the name of the collection to add the content to")
	pageTypes := flag.String("types", "", "the page type to filter by")
	limit := flag.Int("limit", -1, "the max number of fixes to apply")
	flag.Parse()

	return *baseDir, *collectionName, *pageTypes, *limit
}

func parsePageTypes(s string) (map[string]bool, error) {
	if s == "" {
		return nil, errors.New("no page types provided")
	}

	typesRaw := strings.Split(s, ",")
	results := make(map[string]bool, 0)

	for _, v := range typesRaw {
		results[strings.TrimSpace(v)] = true
	}

	return results, nil
}

func errExit(err error) {
	log.Event(nil, "Filter and process script returned an error", log.Error(err))
	os.Exit(1)
}
