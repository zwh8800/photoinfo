package main

import (
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"syscall"

	"time"

	"github.com/rwcarlsen/goexif/exif"
)

func main() {
	dir := os.Args[1]

	filepath.Walk(dir, func(filepath string, info os.FileInfo, err error) error {
		log.Println("filepath: ", filepath)
		file, err := os.Open(filepath)
		if err != nil {
			log.Println("os.Open: ", err)
			return nil
		}
		var datetime time.Time
		needCheckFilename := false
		ex, err := exif.Decode(file)
		if err != nil {
			log.Println("exif.Decode: ", err)
			needCheckFilename = true
		} else {
			datetime, err = ex.DateTime()
			if err != nil {
				log.Println("ex.DateTime: ", err)
				needCheckFilename = true
			}
		}
		if needCheckFilename {
			_, filename := path.Split(filepath)
			if len(filename) < 4 {
				log.Println("filename too short: ", filename)
				return nil
			}
			prefix := filename[:4]
			datestr := ""
			format := "20060102_150405"
			switch prefix {
			case "IMG_":
				fallthrough
			case "VID_":
				datestr = filename[4:19]

			case "SAVE":
				fallthrough
			case "PANO":
				datestr = filename[5:20]

			case "IMG2":
				datestr = filename[3:18]

			case "meip":
				datestr = filename[7:21]
				format = "20060102150405"

			case "Scre":
				datestr = filename[11:30]
				format = "2006-01-02-15-04-05"

			case "micr":
				format = ""
				datestr = filename[9:22]
				fallthrough
			case "mmex":
				format = ""
				datestr = filename[8:21]

				timestamp, err := strconv.ParseInt(datestr, 10, 64)
				if err != nil {
					log.Println("time.Parse: ", err, datestr)
					return nil
				}
				datetime = time.Unix(0, timestamp*1000*1000)
			}
			if format != "" {
				datetime, err = time.Parse(format, datestr)
				if err != nil {
					log.Println("time.Parse: ", err, datestr)
					return nil
				}
			}
		}

		stat, err := file.Stat()
		if err != nil {
			log.Println("file.Stat: ", err)
			return nil
		}
		sysStat, ok := stat.Sys().(*syscall.Stat_t)
		if !ok {
			log.Println("cast to *syscall.Stat_t failed")
			return nil
		}

		atime := time.Unix(sysStat.Atimespec.Unix())
		ctime := time.Unix(sysStat.Ctimespec.Unix())
		mtime := time.Unix(sysStat.Mtimespec.Unix())

		log.Println(datetime, ctime, atime, mtime)

		if err := os.Chtimes(filepath, datetime, datetime); err != nil {
			log.Println("os.Chtimes: ", err)
			return nil
		}

		return nil
	})
}
