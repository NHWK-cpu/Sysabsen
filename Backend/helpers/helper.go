package helpers

import "strings"

func ParseCommaSeparated(s string) []string {
    if s == "" {
        return []string{}
    }
    res := strings.Split(s, ",")
    for i := range res {
        res[i] = strings.TrimSpace(res[i])
    }
    return res
}