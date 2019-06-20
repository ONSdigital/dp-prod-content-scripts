package main

import (
	"errors"
	"flag"
	"os"
	"strings"

	"github.com/ONSdigital/dp-zebedee-utils/content"
	"github.com/ONSdigital/log.go/log"
)

type Counter struct {
	targetTypes map[string]int
	total       int
}

func main() {
	dir := flag.String("dir", "", "the zebedee master dir")
	targetTypes := flag.String("types", "", "comma separated list of page types to count")
	flag.Parse()

	if !content.Exists(*dir) {
		errExit(errors.New("master dir does not exist"))
	}

	if *targetTypes == "" {
		errExit(errors.New("page type not specified"))
	}

	types := strings.Split(*targetTypes, ",")
	typeMap := map[string]int{}
	for _, val := range types {
		typeMap[strings.TrimSpace(val)] = 0
	}

	log.Event(nil, "running count job for pageTypes", log.Data{
		"types": targetTypes,
		"dir":   *dir,
	})
	if err := content.FilterAndProcess(*dir, &Counter{total: 0, targetTypes: typeMap}); err != nil {
		errExit(err)
	}
}

func errExit(err error) {
	log.Event(nil, "Filter and process script returned an error", log.Error(err))
	os.Exit(1)
}

func (c *Counter) Filter(path string, info os.FileInfo) (bool, error) {
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

	if _, ok := c.targetTypes[pageType.Value]; ok {
		return strings.Contains(string(jBytes), "@ons.gsi.gov.uk"), nil
	}

	return false, nil
}

func (c *Counter) Process(path string) error {
	jBytes, err := content.ReadJson(path)
	if err != nil {
		return err
	}

	pageType, err := content.GetPageType(jBytes)
	if err != nil {
		return err
	}

	if count, ok := c.targetTypes[pageType.Value]; ok {
		c.targetTypes[pageType.Value] = count + 1
	} else {
		c.targetTypes[pageType.Value] = 0
	}

	c.total += 1
	return nil
}

func (c *Counter) OnComplete() error {
	log.Event(nil, "count page types contain gsi emails complete", log.Data{
		"page_types": c.targetTypes,
		"total":      c.total,
	})
	return nil
}

func (c *Counter) LimitReached() bool {
	return false
}
