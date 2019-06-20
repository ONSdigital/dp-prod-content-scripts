package collections

import (
	"github.com/ONSdigital/log.go/log"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type ContentMove struct {
	Collection    *Collection
	MovingFromAbs string
	MovingFromRel string
	MovingToRel   string
	MasterDir     string
}

func MoveContent(move ContentMove) (map[string]string, error) {
	// from -> to
	completedMoves := make(map[string]string)

	err := filepath.Walk(move.MovingFromAbs, func(absoluteSrcPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if move.MovingFromAbs == absoluteSrcPath || info.IsDir() {
			return nil
		}

		// relative to the root dir of the content being move.
		uri, _ := filepath.Rel(move.MovingFromAbs, absoluteSrcPath)

		// the taxonomy uri the content is being moved to
		moveToTaxonomyURI := path.Join(move.MovingToRel, uri)

		err = move.Collection.MoveContent(absoluteSrcPath, move.MovingFromRel, moveToTaxonomyURI)
		if err != nil {
			return err
		}

		relSrc, _ := filepath.Rel(move.MasterDir, absoluteSrcPath)
		completedMoves[relSrc] = moveToTaxonomyURI
		return nil
	})
	return completedMoves, err
}

func FindUsesOfUris(p ContentMove) (map[string]string, error) {
	log.Event(nil, "Scanning master for uses of uri", log.Data{"uri": p.MovingFromRel})
	brokenUris := make(map[string]string)

	err := filepath.Walk(p.MasterDir, func(srcFilePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || filepath.Ext(srcFilePath) != ".json" {
			// skip.
			return nil
		}

		b, err := ioutil.ReadFile(srcFilePath)
		if err != nil {
			return err
		}

		fileStr := string(b)
		if strings.Contains(fileStr, p.MovingFromRel) {
			brokenUris[srcFilePath] = srcFilePath
		}
		return nil
	})
	return brokenUris, err
}

func FixUris(p ContentMove, affectedFiles map[string]string, completedMoves map[string]string) ([]string, error) {
	brokenLinks := make([]string, 0)
	for _, srcFilePath := range affectedFiles {
		relURI, err := filepath.Rel(p.MasterDir, srcFilePath)
		if err != nil {
			return nil, err
		}
		_, alreadyMoved := completedMoves[relURI]
		if alreadyMoved {
			continue
		}

		b, err := ioutil.ReadFile(srcFilePath)
		if err != nil {
			return nil, err
		}

		relPath, _ := filepath.Rel(p.MasterDir, srcFilePath)

		if err := p.Collection.AddContent(relPath, FixBrokenLinks(b, p.MovingFromRel, p.MovingToRel)); err != nil {
			return nil, err
		}

		brokenLinks = append(brokenLinks, relURI)
	}
	return brokenLinks, nil
}
