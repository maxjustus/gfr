package main

import (
    "fmt"
    "os"
    "io"
    "bytes" /* use replacer */
    "path/filepath"
    "sync"
)

func main() {
    var waitgroup sync.WaitGroup

    filepath.Walk("./test", func (path string, fileinfo os.FileInfo, err error) error {
        if err != nil { fmt.Println(err) }
        if !fileinfo.IsDir() {
            waitgroup.Add(1)
            go func() {
                ScanFile(path, "Cool", "Kewl")
                waitgroup.Done()
            }()
        }
        return nil
    })

    waitgroup.Wait()
}

func ScanFile(path string, matcher string, replacement string) {
    f, err := os.Open(path)
    if err != nil { panic(err) }
    defer f.Close()

    stat, err := f.Stat()
    if err != nil { panic(err) }
    tempfile, err := os.OpenFile(".replacement", os.O_WRONLY|os.O_APPEND|os.O_CREATE, stat.Mode())
    defer tempfile.Close()
    defer os.Remove(tempfile.Name())
    defer os.Rename(tempfile.Name(), path)

    matchCount := 0
    buf   := make([]byte, 32)
    last  := make([]byte, len(matcher))
    match := make([]byte, 64)

    for {
        n, err := f.Read(buf)
        if err != nil && err != io.EOF { panic(err) }

        chunk := buf[:n]
        fmt.Println(last)

        copyToMatch(match, chunk, last)

        matchCount = bytes.Count(match, []byte(matcher))

        for {
            idx := bytes.Index(match, []byte(matcher))
            copy(match, bytes.Replace(match, []byte(matcher), []byte(replacement), 1))

            if idx == -1 {
                break
            }
        }
        fmt.Printf("matchct %v\n", matchCount)
        fmt.Printf("match %v\n", string(match))

        matchLen := nonNullByteCount(match)
        readAhead := len(matcher)
        if readAhead > n {
            readAhead = n
        }
        offset := matchLen - readAhead
        fmt.Println(offset)

        if matchCount > 0 {
            tempfile.Write(match[:offset])

            copy(last, match[offset:])
        } else {
            if n == 0 {
                tempfile.Write(match[:matchLen])
                break
            } else {
                tempfile.Write(match[:offset])
            }
            copy(last, match[offset:])
        }
    }
}

func copyToMatch(match, chunk, last []byte) {
        i := 0

        for i, _ := range match {
            match[i] = 0
        }

        for _, b := range last {
            if b != 0 {
                match[i] = b
                i++
            }
        }

        for _, b := range chunk {
            if b != 0 {
                match[i] = b
                i++
            }
        }
}

func nonNullByteCount(a []byte) int {
    byteCt := 0
    for _, b := range a {
        if b != 0 {
            byteCt++
        }
    }

    return byteCt
}
