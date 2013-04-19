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
)

func main() {
    runtime.GOMAXPROCS(runtime.NumCPU())

    var waitGroup sync.WaitGroup
    var outMutex sync.Mutex
    throttle := make(chan bool, 60)
    from, to, dir := parseFlags()

    filepath.Walk(dir, func (path string, fileinfo os.FileInfo, err error) error {
        if err != nil { fmt.Println(err) }

        stat, err := os.Lstat(path)
        if err != nil { panic(err) }

        if (stat.Mode() & os.ModeType == 0) && !bytes.Contains([]byte(path), []byte(".git")) {
            waitGroup.Add(1)
            throttle <- true
            go func() {
                ScanFile(path, stat, from, to, outMutex)
                waitGroup.Done()
                <-throttle
            }()
        }
        return nil
    })

    waitGroup.Wait()
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

func ScanFile(path string, stat os.FileInfo, matcher string, replacement string, outMutex sync.Mutex) {
    f, err := os.Open(path)
    if err != nil { panic(err) }
    defer f.Close()

    tempfile, err := ioutil.TempFile("./", "replacement")

    err = os.Chmod(tempfile.Name(), stat.Mode())
    if err != nil { panic(err) }

    defer tempfile.Close()
    defer os.Rename(tempfile.Name(), path)
    r := bufio.NewReader(f)
    pathShown := false
    var lineNumber uint64 = 0

    for {
        line, err := r.ReadBytes('\n')
        if err != nil && err != io.EOF { panic(err) }
        if len(line) == 0 { break }

        lineNumber += 1
        count := bytes.Count(line, []byte(matcher))
        replaced := bytes.Replace(line, []byte(matcher), []byte(replacement), -1)
        if count > 0 {
            //Yucky, need to figure out a way to have a printing goroutine instead of a mutex
            outMutex.Lock()
            if !pathShown {
                pathShown = true
                fmt.Println(path)
            }
            fmt.Printf("-%v: %v+%v: %v\n", lineNumber, string(line), lineNumber, string(replaced))
            outMutex.Unlock()
        }
        tempfile.Write(replaced)
    }
}
