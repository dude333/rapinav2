// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package infra

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

//
// UnzipVerbosity will decompress a zip archive, moving all files and folders
// within the zip file (parameter 1) to an output directory (parameter 2).
// Source: https://golangcode.com/unzip-files-in-go/
//
func Unzip(src, dest string, filters []string, verbose bool) ([]string, error) {

	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {

		// deferInLoop: stackoverflow.com/a/45617791/276311
		ff, err := func() ([]string, error) {
			if !matchFilter(f.Name, filters) {
				return filenames, nil // continue
			}

			rc, err := f.Open()
			if err != nil {
				return filenames, err
			}
			defer rc.Close()

			// Store filename/path for returning and using later on
			fpath, err := sanitizeArchivePath(dest, f.Name)
			if err != nil {
				return filenames, err
			}

			// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
			if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
				return filenames, fmt.Errorf("%s: illegal file path", fpath)
			}

			filenames = append(filenames, fpath)

			if f.FileInfo().IsDir() {

				// Make Folder
				if err := os.MkdirAll(fpath, os.ModePerm); err != nil {
					return nil, err
				}

			} else {

				// Make File
				if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
					return filenames, err
				}

				outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
				if err != nil {
					return filenames, err
				}

				counter := io.Discard
				if verbose {
					fmt.Printf("[          ] Unziping %s", fpath)
					counter = &WriteCounter{}
				}
				_, err = io.Copy(outFile, io.TeeReader(rc, counter))
				if verbose {
					fmt.Println()
				}

				// Close the file without defer to close before next iteration of loop
				outFile.Close()

				if err != nil {
					return filenames, err
				}

			}
			return filenames, err
		}()

		if err != nil {
			return ff, err
		}
	}

	return filenames, nil
}

func matchFilter(s string, f []string) bool {
	if len(f) == 0 {
		return true // ignore when filter f is empty
	}
	for i := range f {
		if strings.Contains(strings.ToLower(s), strings.ToLower(f[i])) {
			return true
		}
	}
	return false
}

// Sanitize archive file pathing from "G305: Zip Slip vulnerability"
func sanitizeArchivePath(d, t string) (v string, err error) {
	v = filepath.Join(d, t)
	if strings.HasPrefix(v, filepath.Clean(d)) {
		return v, nil
	}

	return "", fmt.Errorf("%s: %s", "content filepath is tainted", t)
}
