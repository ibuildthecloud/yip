package plugins

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-multierror"
	"github.com/mudler/yip/pkg/schema"
	log "github.com/sirupsen/logrus"
	"github.com/twpayne/go-vfs"
)

func EnsureDirectories(s schema.Stage, fs vfs.FS, console Console) error {
	var errs error
	for _, dir := range s.Directories {
		if err := writePath(dir, fs, true); err != nil {
			log.Error(err.Error())
			errs = multierror.Append(errs, err)
			continue
		}
	}
	return errs
}

func writeDirectory(dir schema.Directory, fs vfs.FS) error {
	log.Debug("Creating directory ", dir.Path)
	err := fs.Mkdir(dir.Path, os.FileMode(dir.Permissions))
	if err != nil {
		return err
	}

	return fs.Chown(dir.Path, dir.Owner, dir.Group)
}

func writePath(dir schema.Directory, fs vfs.FS, topLevel bool) error {
	inf, err := fs.Stat(dir.Path)
	if err == nil && inf.IsDir() && topLevel {
		// The path already exists, apply permissions and ownership only
		err = fs.Chmod(dir.Path, os.FileMode(dir.Permissions))
		if err != nil {
			return err
		}
		return fs.Chown(dir.Path, dir.Owner, dir.Group)
	} else if err == nil && !inf.IsDir() {
		return fmt.Errorf("Error, '%s' already exists and it is not a directory", dir.Path)
	} else if err == nil {
		return nil
	} else {
		parentDir := filepath.Dir(dir.Path)
		_, err = fs.Stat(parentDir)
		if parentDir == "/" || parentDir == "." || err == nil {
			//There is no parent dir or it already exists
			return writeDirectory(dir, fs)
		} else {
			//Parent dir needs to be created
			pDir := schema.Directory{parentDir, dir.Permissions, dir.Owner, dir.Group}
			err = writePath(pDir, fs, false)
			if err != nil {
				return err
			}
			return writeDirectory(dir, fs)
		}
	}
}
