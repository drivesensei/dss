package main

import "github.com/drivesensei/nicesize"

type ScanResult struct {
	Count int    `json:"count"`
	Size  int64  `json:"size"`
	HSize string `json:"hsize"`
}

func handleScan(data ScanFoldersType) (ScanResult, error) {
	Size := uint64(2500 * 1024)
	HSize := nicesize.Format(Size)
	Count := 2048
	return ScanResult{Count, int64(Size), HSize}, nil
}
