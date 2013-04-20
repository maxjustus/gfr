package main

import (
    "testing"
    "io/ioutil"
    "strings"
    "os"
)

func TestScanFile(t *testing.T) {
    f, err := ioutil.TempFile("./", "test")
    if err != nil { panic(err) }
    defer os.Remove(f.Name())

    original := "herp\nfun\nherpzherpzherp herpzherpzherp lol herp hahahaha herp man herp man dude herp buggz buggz buggz herp dorp herpp\nherp"
    expected := strings.Replace(original, "herp", "0", -1)

    f.WriteString(original)

    stat, err := os.Lstat("./" + f.Name())
    if err != nil { panic(err) }

    ScanFile(f.Name(), stat, "herp", "0")
    changed, err := ioutil.ReadFile(f.Name())
    if string(changed) != expected {
        t.Error("Expected", expected, "got", string(changed))
    }

    if err != nil { panic(err) }

}
