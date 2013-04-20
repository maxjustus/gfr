package main

import (
    "gfr/flags"
    "gfr/scan"
    "fmt"
    "os"
    "path/filepath"
    "sync"
    "runtime"
)

func main() {
    runtime.GOMAXPROCS(runtime.NumCPU())

    var waitGroup sync.WaitGroup
    throttle := make(chan bool, 60)
    out := make(chan []string, 60)
    done := make(chan bool)
    from, to, dir, formats := flags.ParseFlags()

    go func() {
        filepath.Walk(dir, func (path string, fileinfo os.FileInfo, err error) error {
            if err != nil { fmt.Println(err) }

            stat, err := os.Lstat(path)
            if err != nil { panic(err) }

            if (stat.Mode() & os.ModeType == 0) && scan.ShouldScan(path, formats) {
                waitGroup.Add(1)
                throttle <- true
                go func() {
                    out <- scan.ScanFile(path, stat, from, to)
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
