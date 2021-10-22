package frontend

import (
	"io/fs"
	"log"
	"os"
)

var ContentFS fs.FS

func init() {
	var err error
	ContentFS = os.DirFS("./frontend/")
	if err != nil {
		log.Fatal(err)
	}
}
