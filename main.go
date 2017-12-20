package main

import (
	"log"
	"path"
	"path/filepath"

	"github.com/colinmarc/hdfs"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	host    = kingpin.Flag("host", "Hostname + address to target (e.g. localhost:8020).").Required().String()
	fromDir = kingpin.Flag("from", "Root directory to start renaming files without .gz to with .gz.").Required().String()
)

func exitOnErr(desc string, err error) {
	if err != nil {
		log.Fatalf("%s: %s", desc, err)
	}
}

func walkRenameDir(dirPath string, client *hdfs.Client) {
	const gzExt = ".gz"

	fileInfo, err := client.ReadDir(dirPath)
	exitOnErr("hdfs.Client.ReadDir", err)

	for _, f := range fileInfo {
		nextPath := path.Join(dirPath, f.Name())

		if f.IsDir() {
			walkRenameDir(nextPath, client)
		} else if filepath.Ext(f.Name()) != gzExt {
			nextGzPath := nextPath + gzExt
			log.Printf("RENAME %s -> %s", nextPath, nextGzPath)

			// allow error to accumulate
			err = client.Rename(nextPath, nextGzPath)

			if err != nil {
				log.Printf("hdfs.Client.Rename: %s", err)
			}
		}
	}
}

func main() {
	kingpin.Parse()

	client, err := hdfs.New(*host)
	exitOnErr("hdfs.New", err)

	srcStat, err := client.Stat(*fromDir)
	exitOnErr("hdfs.Client.Stat", err)

	if !srcStat.IsDir() {
		log.Fatalf("Given from dir path '%s' is not a directory!", *fromDir)
	}

	// recursive portion
	walkRenameDir(*fromDir, client)
}
