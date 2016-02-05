package main

import (
    "fmt"
    "os"
    "io"

    "github.com/dsoprea/go-exiftool"
    flags "github.com/jessevdk/go-flags"
)

const (
    CopyBlockSize = 8192
    OutputFilepath = "updated.jpg"
    ThumbnailFilepath = "thumb.jpg"
)

type tagVisitor struct {
}

func (tv *tagVisitor) HandleTag(tagName *string, value *string) error {
    fmt.Printf("TAG [%s]=[%s]\n", *tagName, *value)

    return nil
}

type options struct {
    ImageFilepath string  `short:"f" long:"image-filepath" description:"Image file-path" required:"true"`
}

func readOptions () *options {
    o := options {}

    _, err := flags.Parse(&o)
    if err != nil {
        os.Exit(1)
    }

    return &o
}

func copyFile(r io.Reader, w io.Writer) (err error) {
    defer func() {
        if r := recover(); r != nil {
            err = r.(error)
        }
    }()

    buffer := make([]byte, CopyBlockSize)
    i := 0
    for {
        n, err := r.Read(buffer)
        if err != nil && err != io.EOF {
            panic(err)
        }

        if n > 0 {
            _, err := w.Write(buffer[:n])
            if err != nil {
                panic(err)
            }
        }

        if n < CopyBlockSize {
            break
        }

        i += len(buffer)
    }

    return nil
}

func copyToFilepath(r io.Reader, toFilepath *string) (err error) {
    defer func() {
        if r := recover(); r != nil {
            err = r.(error)
        }
    }()

    // Remove the "to" file if it already exists.
    f, err := os.Open(*toFilepath)
    if err == nil {
        f.Close()
        os.Remove(*toFilepath)
    }

    w, err := os.Create(*toFilepath)
    if err != nil {
        f.Close()
        panic(err)
    }

    err = copyFile(r, w)
    if err != nil {
        panic(err)
    }

    w.Close()

    return nil
}

func main() {
    var imageFilepath string

    o := readOptions()
    imageFilepath = o.ImageFilepath

    // Copy the file.

    f, err := os.Open(imageFilepath)
    if err != nil {
        panic(err)
    }

    of := OutputFilepath
    err = copyToFilepath(f, &of)
    f.Close()

    if err != nil {
        panic(err)
    }

    // Dump the tags (using our visitor).

    fmt.Println("Dumping tags.")
    fmt.Println("")

    et := exiftool.NewExifTool(&of)
//    et.SetShowCommands(true)

    tv := &tagVisitor {}
    err = et.ReadTags(tv)
    if err != nil {
        panic(err)
    }

    fmt.Println("")

    // Update tag.

    fmt.Println("Setting tag.")

    err = et.SetTag("GPS", "Longitude", []string { "80", "1", "6", "1", "6", "1" })
    if err != nil {
        panic(err)
    }

    // If there's a thumbnail, write it.

    ht, err := et.HasThumbnail()
    if err != nil {
        panic(err)
    }

    if ht == true {
        r, err := et.GetThumbnail()
        if err != nil {
            panic(err)
        }

        tf := ThumbnailFilepath
        err = copyToFilepath(r, &tf)
        if err != nil {
            panic(err)
        }

        thumbnailTempFilepath := r.Name()
        r.Close()

        os.Remove(thumbnailTempFilepath)

        fmt.Printf("Wrote thumbnail: [%s]\n", tf)
    } else {
        fmt.Printf("No thumbnail.\n")
    }
}
