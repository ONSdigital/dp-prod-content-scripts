package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/ONSdigital/dp-zebedee-utils/content"
	"github.com/ONSdigital/log.go/log"
	"github.com/pkg/errors"
)

const layout = "2006-01-02T15:04:05.000Z"

type CollectionJson struct {
	ApprovalStatus string   `json:"approvalStatus"`
	Events         []*Event `json:"events"`
}

type Event struct {
	Type string `json:"type"`
	Date string `json:"date"`
}

func main() {
	log.Namespace = "approval_times"
	collectionsDir := flag.String("dir", "", "the collections dir")
	collectionName := flag.String("col", "", "the name of the target collection")
	flag.Parse()

	if collectionsDir == nil {
		exit(errors.New("no dir arg was provided"))
	}

	if collectionName == nil {
		exit(errors.New("no col arg was provided"))
	}

	collectionJson := filepath.Join(*collectionsDir, *collectionName+".json")

	if !content.Exists(collectionJson) {
		exit(errors.New("collection json file does not exit"))
	}

	b, err := ioutil.ReadFile(collectionJson)
	if err != nil {
		exit(err)
	}

	var col CollectionJson
	err = json.Unmarshal(b, &col)
	if err != nil {
		exit(err)
	}

	start, end, err := col.getApprovalEvents()
	if err != nil {
		exit(err)
	}

	d, err := getDuration(start, end)
	if err != nil {
		exit(err)
	}

	log.Event(nil, "collection approval stats", log.Data{
		"collection": collectionName,
		"start_time": start.Date,
		"end_time":   end.Date,
		"duration":   d.String(),
	})
}

func (c CollectionJson) getApprovalEvents() (*Event, *Event, error) {
	if c.ApprovalStatus != "COMPLETE" {
		return nil, nil, errors.New("approval not complete")
	}
	var start *Event = nil
	var end *Event = nil

	index := len(c.Events) - 1
	for i := index; i >= 0; i-- {
		e := c.Events[i]
		if e.Type == "APPROVED" {
			end = e
			index = i - 1
			break
		}
	}

	for i := index; i >= 0; {
		e := c.Events[i]
		if e.Type == "APPROVE_SUBMITTED" {
			start = e
			break
		}
	}

	if start == nil || end == nil {
		return nil, nil, errors.New("could not find approval start and end events")
	}

	return start, end, nil
}

func getDuration(start *Event, end *Event) (time.Duration, error) {
	startT, err := time.Parse(layout, start.Date)
	if err != nil {
		return 0, err
	}

	endT, err := time.Parse(layout, end.Date)
	if err != nil {
		return 0, err
	}

	return endT.Sub(startT), nil
}

func exit(err error) {
	log.Event(nil, "app error", log.Error(err))
	os.Exit(1)
}
