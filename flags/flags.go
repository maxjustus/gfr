package flags

import (
    "flag"
    "strings"
)

func ParseFlags() (from, to, dir string, formats []string) {
    var formatsString string
    flag.StringVar(&formatsString, "fmatch", "", "Comma separated list of filename matchers to limit search to")
    flag.StringVar(&formatsString, "f", "", "Comma separated list of filename matchers to limit search to")

    flag.Parse()
    from = flag.Arg(0)
    to   = flag.Arg(1)
    dir  = flag.Arg(2)

    if dir[0:2] != "./" {
        dir = "./" + dir
    }

    formats = strings.Split(formatsString, ",")
    for i, f := range formats {
        formats[i] = strings.TrimSpace(f)
    }

    return
}
