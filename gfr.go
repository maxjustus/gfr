package main

import (
    "fmt"
    "os"
    "io"
    "io/ioutil"
    "bytes"
    "path/filepath"
    "sync"
    "flag"
    "bufio"
    "runtime"
    "unicode/utf8"
)

const CLR_R = "\x1b[31;1m"
const CLR_G = "\x1b[32;1m"
const CLR_N = "\x1b[0m"

func main() {
    runtime.GOMAXPROCS(runtime.NumCPU())

    var waitGroup sync.WaitGroup
    throttle := make(chan bool, 60)
    out := make(chan []string, 60)
    done := make(chan bool)
    from, to, dir := parseFlags()

    go func() {
        filepath.Walk(dir, func (path string, fileinfo os.FileInfo, err error) error {
            if err != nil { fmt.Println(err) }

            stat, err := os.Lstat(path)
            if err != nil { panic(err) }

            if (stat.Mode() & os.ModeType == 0) && !bytes.Contains([]byte(path), []byte(".git")) {
                waitGroup.Add(1)
                throttle <- true
                go func() {
                    out <- ScanFile(path, stat, from, to)
                    waitGroup.Done()
                    <-throttle
                }()
            }
            return nil
        })

        waitGroup.Wait()
        done <- true
    }()

    for {
        select {
        case lines := <-out:
            for _, line := range lines {
                fmt.Println(line)
            }
        case <-done:
            return
        }
    }
}

func parseFlags() (from string, to string, dir string) {
    flag.Parse()
    from = flag.Arg(0)
    to   = flag.Arg(1)
    dir  = flag.Arg(2)
    if dir[0:2] != "./" {
        dir = "./" + dir
    }
    return
}

func ScanFile(path string, stat os.FileInfo, matcher string, replacement string) (diffs []string) {
    f, err := os.Open(path)
    if err != nil { panic(err) }
    defer f.Close()
    r := bufio.NewReader(f)
    formatCheck, err := r.Peek(1024)
    if err != nil && err != io.EOF {
        fmt.Println(err)
        return
    }

    if !isValidFile(formatCheck) {
        return
    }

    tempfile, err := ioutil.TempFile("", "replacement")

    err = os.Chmod(tempfile.Name(), stat.Mode())
    if err != nil { panic(err) }

    defer tempfile.Close()
    defer os.Rename(tempfile.Name(), path)
    pathShown := false
    var lineNumber uint64 = 0

    diffs = make([]string, 0)

    for {
        line, err := r.ReadBytes('\n')
        if err != nil && err != io.EOF { panic(err) }
        if len(line) == 0 { break }

        lineNumber += 1
        count := bytes.Count(line, []byte(matcher))
        replaced := bytes.Replace(line, []byte(matcher), []byte(replacement), -1)

        if count > 0 {
            if !pathShown {
                pathShown = true
                diffs = append(diffs, path)
            }
            diffs = append(diffs, fmt.Sprintf("%v-%v: %v%v+%v: %v%v",
                                               CLR_R,
                                               lineNumber,
                                               string(line),
                                               CLR_G,
                                               lineNumber,
                                               string(replaced),
                                               CLR_N))
        }

        tempfile.Write(replaced)
    }

    return diffs
}

func isValidFile(filePeek []byte) (valid bool) {
    if !utf8.Valid(filePeek) {
        return false
    }
    return true
}
