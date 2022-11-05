package downloader

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/schollz/progressbar/v3"
)

func Get(url, dir string) error {
	file := path.Base(url)
	target := filepath.Join(dir, file)

	tempDestinationPath := target + ".tmp"

	_, err := os.Stat(tempDestinationPath)
	if err == nil {
		err = os.Remove(tempDestinationPath)
		if err != nil {
			return err
		}
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	f, err := os.OpenFile(tempDestinationPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		file,
	)
	_, err = io.Copy(io.MultiWriter(f, bar), resp.Body)
	if err != nil {
		return err
	}

	err = f.Close()
	if err != nil {
		return err
	}

	err = os.Rename(tempDestinationPath, target)
	if err != nil {
		return err
	}

	if filepath.Ext(target) == ".zip" {
		newFolder := strings.TrimRight(target, ".zip")

		err := os.RemoveAll(newFolder)
		if err != nil {
			return err
		}

		if _, err = os.Stat(newFolder); os.IsNotExist(err) {
			err = os.Mkdir(newFolder, 0755)
			if err != nil {
				return err
			}
		}

		unzip(target, newFolder)
	}

	return nil
}

func unzip(src string, destination string) ([]string, error) {

	// a variable that will store any
	//file names available in a array of strings
	var filenames []string

	// OpenReader will open the Zip file
	// specified by name and return a ReadCloser
	// Readcloser closes the Zip file,
	// rendering it unusable for I/O
	// It returns two values:
	// 1. a pointer value to ReadCloser
	// 2. an error message (if any)
	r, err := zip.OpenReader(src)

	// if there is any error then
	// (err!=nill) becomes true
	if err != nil {
		// and this block will break the loop
		// and return filenames gathered so far
		// with an err message, and move
		// back to the main function

		return filenames, err
	}

	defer r.Close()
	// defer makes sure the file is closed
	// at the end of the program no matter what.

	var uncompressedSize64 uint64 = 0
	for _, f := range r.File {
		uncompressedSize64 += f.UncompressedSize64
	}

	bar := progressbar.DefaultBytes(
		int64(uncompressedSize64),
		"unzip "+filepath.Base(src),
	)

	for _, f := range r.File {

		// this loop will run until there are
		// files in the source directory & will
		// keep storing the filenames and then
		// extracts into destination folder until an err arises

		// Store "path/filename" for returning and using later on
		fpath := filepath.Join(destination, f.Name)

		// Checking for any invalid file paths
		if !strings.HasPrefix(fpath, filepath.Clean(destination)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s is an illegal filepath", fpath)
		}

		// the filename that is accessed is now appended
		// into the filenames string array with its path
		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {

			// Creating a new Folder
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Creating the files in the target directory
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}

		// The created file will be stored in
		// outFile with permissions to write &/or truncate
		outFile, err := os.OpenFile(fpath,
			os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
			f.Mode())

		// again if there is any error this block
		// will be executed and process
		// will return to main function
		if err != nil {
			// with filenames gathered so far
			// and err message
			return filenames, err
		}

		rc, err := f.Open()

		// again if there is any error this block
		// will be executed and process
		// will return to main function
		if err != nil {
			// with filenames gathered so far
			// and err message back to main function
			return filenames, err
		}

		_, err = io.Copy(io.MultiWriter(outFile, bar), rc)

		// Close the file without defer so that
		// it closes the outfile before the loop
		// moves to the next iteration. this kinda
		// saves an iteration of memory & time in
		// the worst case scenario.
		outFile.Close()
		rc.Close()

		// again if there is any error this block
		// will be executed and process
		// will return to main function
		if err != nil {
			// with filenames gathered so far
			// and err message back to main function
			return filenames, err
		}
	}

	// Finally after every file has been appended
	// into the filenames string[] and all the
	// files have been extracted into the
	// target directory, we return filenames
	// and nil as error value as the process executed
	// successfully without any errors*
	// *only if it reaches until here.
	return filenames, nil
}
