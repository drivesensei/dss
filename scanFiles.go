package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"sync"
)

type FileOrFolderType struct {
	Modified string `json:"modified"`
	Path     string `json:"path"`
	Size     uint64 `json:"size"`
}

type FileResult struct {
	File   FileOrFolderType `json:"file"`
	Folder string           `json:"folder"`
}

type ScanResult struct {
	Count       uint64 `json:"count"`
	Size        uint64 `json:"size"`
	Dirs        uint64 `json:"dirs"`
	ErrorsCount uint64 `json:"errors"`
}

type ScanFoldersResult struct {
	Folders []FileOrFolderType `json:"folders"`
	Files   []FileResult       `json:"allFiles"`
	Errors  []string           `json:"errors"`
	Stats   ScanResult         `json:"stats"`
}

func handleScan(data ScanFoldersType) (ScanFoldersResult, error) {
	var result ScanFoldersResult

	var mu sync.Mutex
	var wg sync.WaitGroup
	ch := make(chan ScanFoldersResult)

	// scan each folder in separate goroutine
	for _, folder := range data.Folders {
		wg.Add(1)

		// scan folder in different go routine
		go func(f string) {
			defer wg.Done()

			res, err := scanOneFolder(f)
			if err != nil {
				log.Println("Error scanning folder", f)
				log.Println(err)
			}

			ch <- res
		}(folder)
	}

	go func() {
		wg.Wait() // Wait for all scanOneFolder goroutines to finish
		close(ch)
	}()

	// collect results
	for r := range ch {
		mu.Lock()
		result.Folders = append(result.Folders, r.Folders...)
		result.Files = append(result.Files, r.Files...)
		result.Errors = append(result.Errors, r.Errors...)

		result.Stats.Count += r.Stats.Count
		result.Stats.Dirs += r.Stats.Dirs
		result.Stats.ErrorsCount += r.Stats.ErrorsCount
		result.Stats.Size += r.Stats.Size

		mu.Unlock()
	}
	return result, nil
}

func scanOneFolder(folder string) (ScanFoldersResult, error) {
	var result ScanFoldersResult

	fileOrDir, err := os.Stat(folder)
	if err != nil {
		return ScanFoldersResult{}, err
	}

	if !fileOrDir.IsDir() {
		size := uint64(fileOrDir.Size())
		modified := fmt.Sprintf("%d", fileOrDir.ModTime().Unix())

		file := FileOrFolderType{Size: size, Modified: modified, Path: fileOrDir.Name()}
		fileResult := FileResult{File: file, Folder: folder} // Note the change here

		result.Files = append(result.Files, fileResult)
		result.Folders = append(result.Folders, fileResult.File)
		result.Stats.Size += size
		result.Stats.Count++
		return result, nil
	}

	root := folder
	fils := os.DirFS(root)

	err = fs.WalkDir(fils, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Println(err)
			result.Stats.ErrorsCount++
			result.Errors = append(result.Errors, err.Error())
			return nil
		}

		switch d.IsDir() {
		case true:
			result.Stats.Dirs++
		case false:
			result.Stats.Count++
			if fileInfo, err := d.Info(); err == nil {
				size := uint64(fileInfo.Size())
				modified := fmt.Sprintf("%d", fileInfo.ModTime().Unix())

				file := FileOrFolderType{Size: size, Modified: modified, Path: path}
				fileResult := FileResult{File: file, Folder: folder} // Note the change here

				result.Files = append(result.Files, fileResult)
				result.Stats.Size += size

			} else {
				result.Stats.ErrorsCount++
				result.Errors = append(result.Errors, err.Error())
			}
		}
		return nil
	})

	if fileOrDir.IsDir() {
		size := uint64(result.Stats.Size)
		modified := fmt.Sprintf("%d", fileOrDir.ModTime().Unix())
		result.Folders = append(result.Folders, FileOrFolderType{Size: size, Modified: modified, Path: folder})
	}

	return result, err
}
