package scan

import (
    "unicode/utf8"
    "strings"
)

func ShouldScan(path string, formats []string) bool {
    defaultIgnore := []string{".git", ".jpg", ".jpeg", ".png"}
    for _, ignore := range defaultIgnore {
        if strings.Contains(path, ignore) {
            return false
        }
    }

    if len(formats) > 0 {
        for _, format := range formats {
            if strings.Contains(path, format) {
                return true
            }
        }
        return false
    } else {
        return false
    }

    return true
}

func isValidFile(filePeek []byte) (valid bool) {
    if !utf8.Valid(filePeek) {
        return false
    }
    return true
}
