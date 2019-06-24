package main

import (
	"encoding/csv"
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
	fixC.ApprovalStatus = collections.CompleteState
	err := collections.Save(fixC)
	checkError(err)

	allCols, err := collections.GetCollections(collectionsDir)
	checkError(err)

	types := parsePageTypes(pageTypes)
	checkError(err)

	outputPath := filepath.Join(baseDir, "gsi-fixes.csv")

	f, err := os.Create(outputPath)
	checkError(err)
	defer f.Close()

	csvW := csv.NewWriter(f)
	err = csvW.Write([]string{"uri", "page type", "generates PDF"})
	checkError(err)

	job := &fix{
		OutputPath: outputPath,
		Limit:      limit,
		FixCount:   0,
		MasterDir:  masterDir,
		AllCols:    allCols,
		FixC:       fixC,
		Blocked:    make([]string, 0),
		PageTypes:  types,
		CSVW:       csvW,
	}

	err = content.FilterAndProcess(masterDir, job)
	checkError(err)
}

func getFlags() (string, string, string, int) {
	baseDir := flag.String("dir", "", "the zebedee master dir")
	collectionName := flag.String("col", "", "the name of the collection to add the content to")
	pageTypes := flag.String("types", "", "the page type to filter by")
	limit := flag.Int("limit", -1, "the max number of fixes to apply")
	flag.Parse()

	return *baseDir, *collectionName, *pageTypes, *limit
}

func parsePageTypes(s string) map[string]bool {
	results := make(map[string]bool, 0)
	if s == "" {
		return results
	}

	typesRaw := strings.Split(s, ",")
	for _, v := range typesRaw {
		results[strings.TrimSpace(v)] = true
	}

	return results
}

func errExit(err error) {
	log.Event(nil, "Filter and process script returned an error", log.Error(err))
	os.Exit(1)
}

func checkError(err error) {
	if err != nil {
		errExit(err)
	}
}
