package readordie

import (
	"archive/zip"
	"fmt"
	"github.com/cavaliercoder/grab"
	"io"
	"os"
	"path"
	"strings"
)

// DownloadAndZip Downloads and generate a zip of a manga chapter
func DownloadAndZip(manga Manga, chapter Chapter, outputFolder string) error {
	pages, err := chapter.ListPages()
	if err != nil {
		return err
	}

	dir := NewID()
	err = os.Mkdir(dir, 0666)
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	files := make([]string, 0)
	for _, page := range pages {
		ext := strings.Split(path.Ext(page.Link), "?")[0]
		fileName := fmt.Sprintf("%v/%03d%v", dir, page.Number, ext)
		resp, err := grab.Get(fileName, page.Link)
		if err != nil {
			return err
		}
		if err = resp.Err(); err != nil {
			return err
		}

		files = append(files, resp.Filename)
	}

	err = zipFiles(outputFolder, fmt.Sprintf("%v - %v.zip", manga, chapter), files)
	if err != nil {
		return err
	}
	return nil
}

func zipFiles(folder string, filename string, files []string) error {
	err := os.MkdirAll(folder, 0666)
	if err != nil {
		return err
	}

	newZipFile, err := os.Create(fmt.Sprintf("%v/%v", folder, filename))
	if err != nil {
		return err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	// Add files to zip
	for _, file := range files {

		zipfile, err := os.Open(file)
		if err != nil {
			return err
		}
		defer zipfile.Close()

		// Get the file information
		info, err := zipfile.Stat()
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}
		if _, err = io.Copy(writer, zipfile); err != nil {
			return err
		}
	}
	return nil
}
