package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ONSdigital/dp-zebedee-utils/collections"
	"github.com/ONSdigital/dp-zebedee-utils/content"
	"github.com/ONSdigital/log.go/log"
)

type fix struct {
	MasterDir   string
	AllCols     *collections.Collections
	FixC        *collections.Collection
	Limit       int
	FixCount    int
	FixLog      map[string]int
	Blocked     []string
	PageTypes   map[string]bool
	AnyPageType bool
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

	if !strings.Contains(jsonStr, oldEmail) {
		return nil
	}

	uri = "/" + uri
	if blocked, name := f.AllCols.IsBlocked(uri); blocked {
		f.Blocked = append(f.Blocked, fmt.Sprintf("%s:%s", name, uri))
		return nil
	}

	jsonStr = strings.Replace(jsonStr, oldEmail, newEmail, -1)

	//log.Event(nil, "applying content fix", log.Data{"uri": uri})
	if err := f.FixC.AddToReviewed(uri, []byte(jsonStr)); err != nil {
		return err
	}



	pageType, err := content.GetPageType([]byte(jsonStr))
	if err != nil {
		return err
	}

	f.logFix(pageType)
	f.FixCount += 1
	return nil
}

func (f *fix) OnComplete() error {
	logD := log.Data{
		"stats":          f.FixLog,
		"fix_count":      f.FixCount,
		"fix_collection": f.FixC.Name,
		"blocked":        f.Blocked,
	}

	if len(f.Blocked) <= 20 {
		logD["blocked"] = f.Blocked
	} else {
		logD["blocked"] = len(f.Blocked)
	}

	log.Event(nil, "script fixing content completed successfully", logD)
	return nil
}

func (f *fix) LimitReached() bool {
	if f.Limit == -1 {
		return false
	}
	return f.FixCount >= f.Limit
}

func (f *fix) logFix(pageType *content.PageType) {
	if count, ok := f.FixLog[pageType.Value]; ok {
		f.FixLog[pageType.Value] = count + 1
	} else {
		f.FixLog[pageType.Value] = 1
	}
}
