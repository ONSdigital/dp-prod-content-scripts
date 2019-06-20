package content

import (
	"os"
	"path/filepath"
)

type FilterJob interface {
	Filter(path string, info os.FileInfo) (bool, error)

	Process(path string) error

	OnComplete() error

	LimitReached() bool
}

func FilterAndProcess(dir string, job FilterJob) error {
	walkFunc := func(path string, info os.FileInfo, err error) error {
		if job.LimitReached() {
			return LimitReached
		}

		proceed, err := job.Filter(path, info)
		if err != nil {
			return err
		}

		if !proceed {
			return nil
		}
		return job.Process(path)
	}

	if err := filepath.Walk(dir, walkFunc); err != nil {
		if err != LimitReached {
			return err
		}
	}

	return job.OnComplete()
}
