package collections

import (
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"

	"github.com/ONSdigital/dp-zebedee-utils/errs"
	"github.com/ONSdigital/log.go/log"
	"github.com/satori/go.uuid"
)

const (
	NotStartState   ApprovalStatus = "NOT_STARTED"
	InProgressState ApprovalStatus = "IN_PROGRESS"
	CompleteState   ApprovalStatus = "COMPLETE"
	ErrorState      ApprovalStatus = "ERROR"
)

type ApprovalStatus string

type Metadata struct {
	Name           string
	CollectionRoot string
	InProgress     string
	Complete       string
	Reviewed       string
	CollectionJSON string
}

type Collection struct {
	Metadata              *Metadata      `json:"-"`
	ApprovalStatus        ApprovalStatus `json:"approvalStatus"`
	PublishComplete       bool           `json:"publishComplete"`
	IsEncrypted           bool           `json:"isEncrypted"`
	CollectionOwner       string         `json:"collectionOwner"`
	TimeSeriesImportFiles []string       `json:"timeseriesImportFiles"`
	ID                    string         `json:"id"`
	Name                  string         `json:"name"`
	Type                  string         `json:"type"`
	Teams                 []interface{}  `json:"teams"`
}

type Collections struct {
	Collections []*Collection
}

func (c *Collections) IsBlocked(uri string) (bool, string) {
	for _, col := range c.Collections {
		if blocked := col.Contains(uri); blocked {
			log.Event(nil, "content exists in another collection", log.Data{
				"collection": col.Name,
				"uri":        uri,
			})
			return true, col.Name
		}
	}
	return false, ""
}

func (c *Collections) GetByName(name string) (*Collection, error) {
	for _, col := range c.Collections {
		if col.Name == name {
			return col, nil
		}
	}
	return nil, errs.New("collection not found", nil, log.Data{"collection": name})
}

func (c *Collections) Add(col *Collection) {
	if col != nil {
		c.Collections = append(c.Collections, col)
	}
}

func New(rootPath string, name string) *Collection {
	newID, _ := uuid.NewV4()
	id := fmt.Sprintf("%s-%s", name, newID.String())
	metadata := NewMetadata(rootPath, name)

	return &Collection{
		Metadata:              metadata,
		ApprovalStatus:        NotStartState,
		CollectionOwner:       "PUBLISHING_SUPPORT",
		IsEncrypted:           false,
		PublishComplete:       false,
		Type:                  "manual",
		ID:                    id,
		Name:                  name,
		TimeSeriesImportFiles: []string{},
		Teams:                 []interface{}{},
	}
}

func NewMetadata(collectionsDirPath string, name string) *Metadata {
	collectionRoot := path.Join(collectionsDirPath, name)

	return &Metadata{
		CollectionRoot: collectionRoot,
		Name:           name,
		CollectionJSON: path.Join(collectionsDirPath, name+".json"),
		InProgress:     path.Join(collectionRoot, "inprogress"),
		Complete:       path.Join(collectionRoot, "complete"),
		Reviewed:       path.Join(collectionRoot, "reviewed"),
	}
}

func (c *Collection) GetName() string {
	return c.Name
}

func (c *Collection) GetRootPath() string {
	return c.Metadata.CollectionRoot
}

func (c *Collection) GetInProgress() string {
	return c.Metadata.InProgress
}

func (c *Collection) GetComplete() string {
	return c.Metadata.Complete
}

func (c *Collection) GetReviewed() string {
	return c.Metadata.Reviewed
}

func (c *Collection) Contains(uri string) bool {
	if Exists(path.Join(c.Metadata.InProgress, uri)) {
		log.Event(nil, "collection contains uri", log.Data{
			"uri":        uri,
			"collection": c.Name,
			"dir":        c.Metadata.InProgress,
		})
		return true
	}

	if Exists(path.Join(c.Metadata.Complete, uri)) {
		log.Event(nil, "collection contains uri", log.Data{
			"uri":        uri,
			"collection": c.Name,
			"dir":        c.Metadata.Complete,
		})
		return true
	}

	if Exists(path.Join(c.Metadata.Reviewed, uri)) {
		log.Event(nil, "collection contains uri", log.Data{
			"uri":        uri,
			"collection": c.Name,
			"dir":        c.Metadata.Reviewed,
		})
		return true
	}
	return false
}

func (c *Collection) AddContent(uri string, fileBytes []byte) error {
	collectionURI := c.inProgressURI(uri)
	return writeContent(collectionURI, fileBytes)
}

func (c *Collection) AddToReviewed(uri string, fileBytes []byte) error {
	collectionURI := c.reviewedURI(uri)
	return writeContent(collectionURI, fileBytes)
}

func (c *Collection) MoveContent(absoluteSrcPath string, relSrcPath string, relDestUri string) error {
	absoluteDest := c.inProgressURI(relDestUri)

	// if not a .json file just copy it into the new location.
	if filepath.Ext(absoluteSrcPath) != ".json" {
		return moveContent(absoluteSrcPath, absoluteDest)
	}

	// otherwise we have to read the file into memory so we can check if we need fix any broken links before moving it
	// to its new location.
	b, err := ioutil.ReadFile(absoluteSrcPath)
	if err != nil {
		return err
	}
	return writeContent(absoluteDest, FixBrokenLinks(b, relSrcPath, relDestUri))
}

func FixBrokenLinks(fileBytes []byte, old string, new string) []byte {
	fileStr := string(fileBytes)
	if !strings.Contains(fileStr, old) {
		return fileBytes
	}
	fileStr = strings.Replace(fileStr, old, new, -1)
	return []byte(fileStr)
}

func (c *Collection) inProgressURI(taxonomyURI string) string {
	return path.Join(c.Metadata.InProgress, taxonomyURI)
}

func (c *Collection) reviewedURI(taxonomyURI string) string {
	return path.Join(c.Metadata.Reviewed, taxonomyURI)
}

func (c *Collection) getDirs() []string {
	return []string{
		c.Metadata.CollectionRoot,
		c.Metadata.InProgress,
		c.Metadata.Complete,
		c.Metadata.Reviewed,
	}
}
