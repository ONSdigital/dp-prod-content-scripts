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
	any       bool
	typeCount map[string]int
	total     int
}

func main() {
	dir := flag.String("dir", "", "the zebedee master dir")
	anyType := flag.Bool("any", false, "")
	targetTypes := flag.String("types", "", "comma separated list of page types to count")
	flag.Parse()

	if !content.Exists(*dir) {
		errExit(errors.New("master dir does not exist"))
	}

	c := &Counter{total: 0, any: *anyType, typeCount: make(map[string]int)}

	if c.any {
		log.Event(nil, "running count job for any page type", log.Data{
			"any": c.any,
			"dir": *dir,
		})

	} else {
		types := strings.Split(*targetTypes, ",")
		for _, val := range types {
			c.typeCount[strings.TrimSpace(val)] = 0
		}
		log.Event(nil, "running count job for pageTypes", log.Data{"types": targetTypes, "dir": *dir})
	}

	if !*anyType && *targetTypes == "" {
		errExit(errors.New("page type not specified"))
	}

	if err := content.FilterAndProcess(*dir, c); err != nil {
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

	if c.any {
		return strings.Contains(string(jBytes), "@ons.gsi.gov.uk"), nil
	}

	if _, ok := c.typeCount[pageType.Value]; ok {
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

	if count, ok := c.typeCount[pageType.Value]; ok {
		c.typeCount[pageType.Value] = count + 1
	} else {
		c.typeCount[pageType.Value] = 0
	}

	c.total += 1
	return nil
}

func (c *Counter) OnComplete() error {
	log.Event(nil, "count page types contain gsi emails complete", log.Data{
		"page_types": c.typeCount,
		"total":      c.total,
	})
	return nil
}

func (c *Counter) LimitReached() bool {
	return false
}
