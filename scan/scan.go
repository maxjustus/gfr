package scan

import (
    "fmt"
    "io"
    "io/ioutil"
    "bytes"
    "os"
    "bufio"
)

const CLR_R = "\x1b[31;1m"
const CLR_G = "\x1b[32;1m"
const CLR_N = "\x1b[0m"

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
