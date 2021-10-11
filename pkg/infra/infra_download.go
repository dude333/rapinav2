// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package infra

import (
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/dustin/go-humanize"
)

func DownloadAndUnzip(url, zip string) ([]string, error) {
	verbose := true

	err := downloadFile(url, zip, verbose)
	if err != nil {
		return []string{}, err
	}

	dataDir := filepath.Dir(zip)

	// Unzip and list files
	files, err := Unzip(zip, dataDir, verbose)
	os.Remove(zip)
	if err != nil {
		return []string{}, err
	}

	return files, nil
}

//
// downloadFile source: https://stackoverflow.com/a/33853856/276311
//
func downloadFile(url, filepath string, verbose bool) error {
	// Create dir if necessary
	basepath := path.Dir(filepath)
	if err := os.MkdirAll(basepath, os.ModePerm); err != nil {
		return err
	}

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}

	// https://www.joeshaw.org/dont-defer-close-on-writable-files/
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()

	// Custom Client
	tr := &http.Transport{
		IdleConnTimeout: _http_timeout,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	// Get the data
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Write the body to file
	counter := ioutil.Discard
	if verbose {
		counter = &WriteCounter{}
	}
	_, err = io.Copy(out, io.TeeReader(resp.Body, counter))
	if err != nil {
		return err
	}

	return err
}

// WriteCounter counts the number of bytes written the io.Writer.
// source: https://golangcode.com/download-a-file-with-progress/
type WriteCounter struct {
	Total uint64
}

// Write implements the io.Writer interface and will be passed to io.TeeReader().
func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += uint64(n)
	wc.printProgress()
	return n, nil
}

func (wc WriteCounter) printProgress() {
	fmt.Printf("\r[  %7s", humanize.Bytes(wc.Total))
}

// Cleanup remove files and return a list of files NOT removed.
func Cleanup(files []string) []string {
	var fails []string

	for _, f := range files {
		if err := os.Remove(f); err != nil {
			fails = append(fails, f)
		}
	}

	return fails
}
