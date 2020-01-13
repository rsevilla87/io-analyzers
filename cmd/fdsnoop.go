package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"syscall"
)

const (
	stdin  string = "0"
	stdout string = "1"
	stderr string = "2"
)

var fdFlags = map[int]string{
	syscall.O_RDONLY:    "O_RDONLY",
	syscall.O_WRONLY:    "O_WRONLY",
	syscall.O_CREAT:     "O_CREAT",
	syscall.O_EXCL:      "O_EXCL",
	syscall.O_NOCTTY:    "O_NOCTTY",
	syscall.O_TRUNC:     "O_TRUNC",
	syscall.O_APPEND:    "O_APPEND",
	syscall.O_NONBLOCK:  "O_NONBLOCK",
	syscall.O_DSYNC:     "O_DSYNC",
	syscall.O_SYNC:      "O_SYNC",
	syscall.O_ASYNC:     "O_ASYNC",
	syscall.O_NOFOLLOW:  "O_NOFOLLOW",
	syscall.O_DIRECT:    "O_DIRECT",
	syscall.O_DIRECTORY: "O_DIRECTORY",
	syscall.O_NOATIME:   "O_NOATIME",
	syscall.O_CLOEXEC:   "O_CLOEXEC",
}

func exitError(err error) {
	fmt.Println(err)
	os.Exit(1)
}

func getFields(l []byte) int64 {
	flagsSt := strings.TrimPrefix(string(l), "flags:")
	flagsSt = strings.TrimSpace(flagsSt)
	flags, err := strconv.ParseInt(flagsSt, 8, 32)
	if err != nil {
		exitError(err)
	}
	return flags
}

func readFlags(p string) (int, error) {
	f, err := os.Open(p)
	if err != nil {
		return 0, err
	}
	buf := bufio.NewReader(f)
	for {
		l, _, err := buf.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			exitError(err)
		}
		if strings.HasPrefix(string(l), "flags") {
			d := getFields(l)
			return int(d), nil
		}
	}
	return 0, nil
}

func main() {
	var flags []string
	pidFlag := flag.Int("pid", 0, "PID to analyze")
	flag.Parse()
	if *pidFlag == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}
	pid := strconv.Itoa(*pidFlag)
	fdPath := path.Join("/proc", string(pid), "fd")
	files, err := ioutil.ReadDir(fdPath)
	if err != nil {
		exitError(err)
	}
	for _, f := range files {
		flags = nil
		if f.Name() == stdin || f.Name() == stdout || f.Name() == stderr {
			continue
		}
		realFile, err := os.Readlink(path.Join(fdPath, f.Name()))
		if err != nil {
			exitError(err)
		}
		_, err = os.Stat(realFile)
		if os.IsNotExist(err) {
			continue
		}
		fmt.Printf("FD: %s\nPath: %s\n", f.Name(), realFile)
		f, err := readFlags(path.Join("/proc", pid, "fdinfo", f.Name()))
		if err != nil {
			exitError(err)
		}
		for fdFlag, desc := range fdFlags {
			// Bitwise and with flags
			if fdFlag&f > 0 {
				flags = append(flags, desc)
			}
		}
		fmt.Printf("Flags: %v\n", flags)
	}
}
