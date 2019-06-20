package collections

import (
	"bufio"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/ONSdigital/dp-zebedee-utils/errs"
	"github.com/ONSdigital/log.go/log"
)

const filePerm = 0755

func Exists(filePath string) bool {
	_, err := os.Stat(filePath)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}

// Save a collection
func Save(c *Collection) error {
	if Exists(c.Metadata.CollectionRoot) {
		return errs.New("cannot create collection as a collection with this name already exists", nil, log.Data{
			"name": c.Name,
			"dir":  c.Metadata.CollectionRoot,
		})
	}

	if err := createCollectionDirectories(c); err != nil {
		return errs.New("error creating collection directories", err, log.Data{"name": c.Name})
	}

	if err := createCollectionJson(c); err != nil {
		return errs.New("error creating collection json", err, log.Data{"name": c.Name})
	}
	log.Event(nil, "collection created successfully", log.Data{"collection": c.Name})
	return nil
}

// Delete a collection
func Delete(rootPath string, name string) error {
	target := path.Join(rootPath, name)
	if !Exists(target) {
		return errs.New("cannot delete collection as it does not exist", nil, log.Data{"collection": name})
	}

	log.Event(nil, "deleting collection", log.Data{"collection": target})
	return os.RemoveAll(target)
}

// Get a collection by collection.description.name
func GetCollection(collectionsRoot string, name string) (*Collection, error) {
	metadata := NewMetadata(collectionsRoot, name)
	if !Exists(metadata.CollectionJSON) {
		log.Event(nil, "no collection json file exists for collection", log.Data{"collection": name})
		return nil, nil
	}

	b, err := ioutil.ReadFile(metadata.CollectionJSON)
	if err != nil {
		return nil, err
	}

	var col Collection
	if err := json.Unmarshal(b, &col); err != nil {
		return nil, err
	}
	col.Metadata = metadata
	return &col, nil
}

// Get all collections.
func GetCollections(collectionsRoot string) (*Collections, error) {
	log.Event(nil, "loading existing collections")
	collectionFiles, err := ioutil.ReadDir(collectionsRoot)
	if err != nil {
		return nil, errs.New("failed to read collections dir", err, nil)
	}

	collections := &Collections{Collections: make([]*Collection, 0)}

	for _, f := range collectionFiles {
		if f.IsDir() {
			c, err := GetCollection(collectionsRoot, f.Name())
			if err != nil {
				return nil, errs.New("failed to load collection", err, log.Data{"collectionName": f.Name()})
			}
			collections.Add(c)
		}
	}
	return collections, nil
}

func GetCollectionContaining(relURI string, cols *Collections) *Collection {
	for _, c := range cols.Collections {
		if c.Contains(relURI) {
			return c
		}
	}
	return nil
}

func writeContent(uri string, fileBytes []byte) error {
	dirs, _ := filepath.Split(uri)

	if err := os.MkdirAll(dirs, filePerm); err != nil {
		return err
	}
	return ioutil.WriteFile(uri, fileBytes, filePerm)
}

func moveContent(srcFilePath string, collectionURI string) error {
	dirs, _ := filepath.Split(collectionURI)

	if err := os.MkdirAll(dirs, filePerm); err != nil {
		return err
	}

	destFile, err := os.Create(collectionURI)
	if err != nil {
		return err
	}
	defer destFile.Close()

	srcFile, err := os.Open(srcFilePath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	data := log.Data{"from": srcFilePath, "to": destFile.Name()}
	copied, err := io.Copy(destFile, srcFile)
	if err != nil {
		return errs.New("failed to copy", err, data)
	}

	info, err := srcFile.Stat()
	if err != nil {
		return err
	}
	if copied != info.Size() {
		data["expected"] = info.Size()
		data["actual"] = copied
		return errs.New("move content failure: copied bytes did not match the expected", nil, data)
	}
	return nil
}

func createCollectionDirectories(c *Collection) error {
	for _, d := range c.getDirs() {
		if err := os.MkdirAll(d, filePerm); err != nil {
			return err
		}
	}
	return nil
}

func createCollectionJson(c *Collection) error {
	f, err := os.Create(c.Metadata.CollectionJSON)
	if err != nil {
		return err
	}

	defer func() {
		if err := f.Close(); err != nil {
			log.Event(nil, "failed to close collection json file", log.Error(err))
			panic(err)
		}
	}()

	b, err := json.MarshalIndent(c, "", "	")
	if err != nil {
		return err
	}

	w := bufio.NewWriter(f)
	w.Write(b)
	w.Flush()
	return nil
}
