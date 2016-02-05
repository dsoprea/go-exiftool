package main

import (
    "fmt"
    "os"

    "github.com/dsoprea/go-exiftool"
)

type tagVisitor struct {
}

func (tv *tagVisitor) HandleTag(tagName *string, value *string) error {
    fmt.Printf("TAG [%s]=[%s]\n", *tagName, *value)

    return nil
}

func main() {
    defer func() {
        if r := recover(); r != nil {
            err := r.(error)
            fmt.Printf("ERROR: %s\n", err.Error())

            os.Exit(1)
        }
    }()

    filepath := "test/20160118_203623_Auburn Ct.jpg"
    et := exiftool.NewExifTool(&filepath)

    tv := &tagVisitor {}
    err := et.ReadTags(tv)
    if err != nil {
        panic(err)
    }

    err = et.SetTag("GPS", "GPSLongitude", []string { "80", "1", "2", "1", "3", "1" }, "updated.jpg")
    if err != nil {
        panic(err)
    }
}
