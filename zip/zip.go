package zip

import (
	"archive/zip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/inimei/backup/log"
)

const (
	singleFileByteLimit = 107374182400 // 1 GB
	chunkSize           = 4096         // 4 KB
)

func copyContents(r io.Reader, w io.Writer) error {
	var size int64
	b := make([]byte, chunkSize)
	for {
		size += chunkSize
		if size > singleFileByteLimit {
			log.Error("file too large, please contact us for assistance")
			return errors.New("file too large, please contact us for assistance")
		}

		length, err := r.Read(b[:cap(b)])
		if err != nil {
			if err != io.EOF {
				log.Error("read file error: %v", err.Error())
				return err
			}
			if length == 0 {
				break
			}
		}

		_, err = w.Write(b[:length])
		if err != nil {
			log.Error("write data failed: %v", err.Error())
			return err
		}
	}

	return nil
}

type zipper struct {
	srcFolder string
	skip      []string
	destFile  string
	writer    *zip.Writer
}

func (z *zipper) shouldSkip(path string) bool {
	if z.skip == nil {
		return false
	}

	//use patten
	for _, p := range z.skip {
		if strings.HasPrefix(path, p) {
			return true
		}
	}

	return false
}

func (z *zipper) zipFile(path string, f os.FileInfo, err error) error {
	if err != nil {
		log.Error(err.Error())
		return err
	}

	//os.PathSeparator
	path = strings.Replace(path, "\\", "/", -1)

	fileName := strings.TrimPrefix(path, z.srcFolder+"/")
	if z.shouldSkip(fileName) {
		log.Info("skip file [%v]", fileName)
		return nil
	}

	// only zip files (directories are created by the files inside of them)
	// TODO allow creating folder when no files are inside
	if !f.Mode().IsRegular() || f.Size() == 0 {
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	log.Info("fileName: %v", fileName)

	w, err := z.writer.Create(fileName)
	if err != nil {
		return err
	}
	// copy contents of the file to the zip writer
	err = copyContents(file, w)
	if err != nil {
		return err
	}
	return nil
}

func (z *zipper) zipFolder() error {
	zipFile, err := os.Create(z.destFile)
	if err != nil {
		log.Error("create dest file [%v] failed: %v", z.destFile, err.Error())
		return err
	}
	defer zipFile.Close()

	z.writer = zip.NewWriter(zipFile)
	err = filepath.Walk(z.srcFolder, z.zipFile)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	err = z.writer.Close()
	if err != nil {
		log.Error(err.Error())
		return err
	}

	return nil
}

// ZipFolder zips the given folder to the a zip file
// with the given name
func ZipFolder(srcFolder string, destFile string, skip []string) error {
	srcFolder = strings.TrimSuffix(srcFolder, "/")
	srcFolder = strings.TrimSuffix(srcFolder, "\\")
	z := &zipper{
		srcFolder: srcFolder,
		destFile:  destFile,
		skip:      skip,
	}
	return z.zipFolder()
}
