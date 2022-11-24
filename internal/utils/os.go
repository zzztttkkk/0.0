package utils

import "os"

func FsExists(fp string) bool {
	_, e := os.Stat(fp)
	return !os.IsNotExist(e)
}
