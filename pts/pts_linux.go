//build +linux
package pts

import (
	"io"
	"io/ioutil"
	"os"
	"path"
	"syscall"
)

type pts struct {
}

func closeAccessible(files []*os.File) {
	for _, file := range files {
		file.Close()
	}
}

func openAccessible(dir string) ([]*os.File, error) {
	var ret []*os.File
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		mode := file.Mode()
		if mode&os.ModeDevice == 0 {
			continue
		}
		if mode&os.ModeCharDevice == 0 {
			continue
		}
		if (int)(file.Sys().(*syscall.Stat_t).Uid) != os.Getuid() {
			continue
		}
		desc, err := os.OpenFile(path.Join(dir, file.Name()), os.O_RDWR, 0600)
		if err != nil {
			continue
		}
		ret = append(ret, desc)
	}
	return ret, nil
}

func (v *pts) Write(p []byte) (n int, err error) {
	files, err := openAccessible("/dev/pts/")
	if err != nil {
		return 0, err
	}
	defer closeAccessible(files)

	N := 0
	for _, file := range files {
		n, err := file.Write(p)
		if err != nil {
			return 0, err
		}
		N += n
	}
	return N, nil
}

func NewPts() io.Writer {
	ret := &pts{}
	return ret
}
