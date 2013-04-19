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
)

func main() {
    var waitgroup sync.WaitGroup
    throttle := make(chan bool, 40)
    from, to, dir := parseFlags()

    filepath.Walk(dir, func (path string, fileinfo os.FileInfo, err error) error {
        if err != nil { fmt.Println(err) }

        stat, err := os.Lstat(path)
        if err != nil { panic(err) }

        if (stat.Mode() & os.ModeType == 0) && !bytes.Contains([]byte(path), []byte(".git")) {
            waitgroup.Add(1)
            throttle <- true
            go func() {
                ScanFile(path, stat, from, to)
                waitgroup.Done()
                <-throttle
            }()
        }
        return nil
    })

    waitgroup.Wait()
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

func ScanFile(path string, stat os.FileInfo, matcher string, replacement string) {
    f, err := os.Open(path)
    if err != nil { panic(err) }
    defer f.Close()

    tempfile, err := ioutil.TempFile("./", "replacement")

    err = os.Chmod(tempfile.Name(), stat.Mode())
    if err != nil { panic(err) }

    fmt.Println(path)
    defer tempfile.Close()
    defer os.Rename(tempfile.Name(), path)
    r := bufio.NewReader(f)

    for {
        line, err := r.ReadBytes('\n')
        if err != nil && err != io.EOF { panic(err) }
        if len(line) == 0 { break }

        replaced := bytes.Replace(line, []byte(matcher), []byte(replacement), -1)
        tempfile.Write(replaced)
    }
}
