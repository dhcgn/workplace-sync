package downloader

import (
	"archive/zip"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dhcgn/workplace-sync/config"
	"github.com/schollz/progressbar/v3"
)

func adjustOverwriteFilesNames(target string, filter map[string]string) (string, error) {
	if filter == nil {
		return target, nil
	}

	for regex, new := range filter {
		r, err := regexp.Compile(regex)
		if err != nil {
			return "", err
		}

		if r.MatchString(filepath.Base(target)) {
			target = filepath.Join(filepath.Dir(target), new)
			return target, nil
		}
	}
	return target, nil
}

func Get(link config.Link, dir string) (string, error) {
	req, err := http.NewRequest("GET", link.Url, nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("%v: status code %d for %v", link.GetDisplayName(), resp.StatusCode, link.Url)
	}

	file := path.Base(link.Url)

	// shot urls are redirected to the actual file https://app/latest -> https://app/1.0.0.zip
	if resp.Request != nil && resp.Request.Response != nil && path.Ext(link.Url) == "" {
		r := resp.Request.Referer()
		if r != "" && strings.HasSuffix(r, ".exe") {
			file = path.Base(r)
		} else {
			l, err := resp.Request.Response.Location()
			if err == nil {
				file = path.Base(l.String())
			}
		}
	}

	target := filepath.Join(dir, file)

	if link.Type == "installer" {
		installerFolder := filepath.Join(dir, "installer")
		if _, err := os.Stat(installerFolder); os.IsNotExist(err) {
			err := os.Mkdir(installerFolder, 0755)
			if err != nil {
				return "", err
			}
		}
		target = filepath.Join(installerFolder, file)
	}

	target, err = adjustOverwriteFilesNames(target, link.OverwriteFilesNames)
	if err != nil {
		return "", err
	}

	tempDestinationPath := target + ".tmp"

	_, err = os.Stat(tempDestinationPath)
	if err == nil {
		err = os.Remove(tempDestinationPath)
		if err != nil {
			return "", err
		}
	}

	f, err := os.OpenFile(tempDestinationPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return "", err
	}

	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		file,
	)
	hash := sha256.New()
	_, err = io.Copy(io.MultiWriter(f, bar, hash), resp.Body)
	if err != nil {
		return "", err
	}

	err = f.Close()
	if err != nil {
		return "", err
	}

	sum := fmt.Sprintf("%x", hash.Sum(nil))
	if link.Hash != "" && link.Hash != sum {
		// TODO handle this error
		_ = os.Remove(tempDestinationPath)
		return sum, fmt.Errorf("%v: hash mismatch for %v. Expected %v, actual %v", link.GetDisplayName(), link.Url, link.Hash[0:12], sum[0:12])
	}

	err = os.Rename(tempDestinationPath, target)
	if err != nil {
		return "", err
	}

	if filepath.Ext(target) == ".zip" {
		decompressFolder := ""
		if link.DecompressFlat {
			decompressFolder = dir
		} else {
			decompressFolder = strings.TrimRight(target, ".zip")

			err := os.RemoveAll(decompressFolder)
			if err != nil {
				return "", err
			}

			if _, err = os.Stat(decompressFolder); os.IsNotExist(err) {
				err = os.Mkdir(decompressFolder, 0755)
				if err != nil {
					return "", err
				}
			}
		}

		_, err = unzip(target, decompressFolder, link)
		if err != nil {
			return "", err
		}

		os.Remove(target)
	}

	return sum, nil
}

func unzip(src string, destination string, link config.Link) ([]string, error) {

	var regex *regexp.Regexp
	if link.DecompressFilter != "" {
		r, err := regexp.Compile(link.DecompressFilter)
		if err == nil {
			regex = r
		}
	}

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

	var files = 0
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		if regex != nil && !regex.MatchString(f.Name) {
			continue
		}
		files += 1
	}

	barFiles := progressbar.Default(
		int64(files),
		"unzip "+filepath.Base(src),
	)

	for _, f := range r.File {
		if f.FileInfo().IsDir() && link.DecompressFlat {
			continue
		}

		if regex != nil && !regex.MatchString(f.Name) {
			continue
		}

		barFiles.Describe("unzip " + filepath.Base(f.Name))
		barFiles.Add(1)

		// this loop will run until there are
		// files in the source directory & will
		// keep storing the filenames and then
		// extracts into destination folder until an err arises

		// Store "path/filename" for returning and using later on
		fpath := ""
		if link.DecompressFlat {
			fpath = filepath.Join(destination, filepath.Base(f.Name))
		} else {
			fpath = filepath.Join(destination, f.Name)
		}

		fpath, err = adjustOverwriteFilesNames(fpath, link.OverwriteFilesNames)
		if err != nil {
			return filenames, err
		}

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

		_, err = io.Copy(outFile, rc)

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
