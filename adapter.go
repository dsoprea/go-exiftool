package exiftool

import (
    "fmt"
    "bytes"
    "strings"
    "os"

    "io/ioutil"
    "os/exec"

    "github.com/dsoprea/go-xmlvisitor"
)

const (
    // The name of the *exif* command-line tool.
    ExifToolFilename = "exif"
    ThumbnailExistsLinePrefix = "EXIF data contains a thumbnail"
    ThumbnailSearchTailLineCount = 5
)

// Whoever calls ReadTags() must provide an implementation of this.
type ExifVisitor interface {
    HandleTag(tagName *string, value *string) error
}

// A visitor that satisfies the "simple" visitor interface for *go-xmlvisitor*.
type xmlVisitor struct {
    ev ExifVisitor
}

func (xv *xmlVisitor) HandleStart(tagName *string, attrp *map[string]string, xp *xmlvisitor.XmlParser) error {
    return nil
}

func (xv *xmlVisitor) HandleEnd(tagName *string, xp *xmlvisitor.XmlParser) error {
    return nil
}

func (xv *xmlVisitor) HandleValue(tagName *string, value *string, xp *xmlvisitor.XmlParser) error {
    return xv.ev.HandleTag(tagName, value)
}

func newXmlVisitor(ev ExifVisitor) (*xmlVisitor) {
    return &xmlVisitor {
        ev: ev,
    }
}


type ExifTool struct {
    filepath string
    showCommands bool
}

func NewExifTool(filepath *string) *ExifTool {
    return &ExifTool {
            filepath: *filepath,
            showCommands: false,
    }
}

func (et *ExifTool) SetShowCommands(showCommands bool) {
    et.showCommands = showCommands
}

func (et *ExifTool) getBytesFromCommand(cmd string, args []string) (buffer *bytes.Buffer, err error) {
    defer func() {
        if r := recover(); r != nil {
            buffer = nil
            err = r.(error)
        }
    }()

    if et.showCommands == true {
        fmt.Printf("CMD: [%s] %s\n", cmd, args)
    }

    // It's necessary to pass in at least one non-variadic argument (otherwise 
    // you'll get warned about there not being enough arguments).
    c := exec.Command(cmd, args...)
    
    var output bytes.Buffer
    c.Stdout = &output
    c.Stderr = &output

    err = c.Run()
    if err != nil {
        fmt.Println(output.String())
        panic(err)
    }

    return &output, nil
}

func (et *ExifTool) getLinesFromCommand(cmd string, args []string) (lines []string, err error) {
    defer func() {
        if r := recover(); r != nil {
            lines = nil
            err = r.(error)
        }
    }()

    if et.showCommands == true {
        fmt.Printf("CMD: [%s] %s\n", cmd, args)
    }

    // It's necessary to pass in at least one non-variadic argument (otherwise 
    // you'll get warned about there not being enough arguments).
    c := exec.Command(cmd, args...)
    
    var output bytes.Buffer
    c.Stdout = &output
    c.Stderr = &output

    err = c.Run()
    if err != nil {
        fmt.Println(output.String())
        panic(err)
    }

    lines = strings.Split(output.String(), "\n")
    return lines, nil
}

// Invoke the one visitor callback with each tag found in the image.
func (et *ExifTool) ReadTags(ev ExifVisitor) (err error) {
    defer func() {
        if r := recover(); r != nil {
            err = r.(error)
        }
    }()

    cmd := []string { "-x", et.filepath }
    outputXml, err := et.getBytesFromCommand(ExifToolFilename, cmd)
    if err != nil {
        panic(err)
    }

    s := outputXml.String()
    r := strings.NewReader(s)
    xv := newXmlVisitor(ev)

    p := xmlvisitor.NewXmlParser(r, xv)

    err = p.Parse()
    if err != nil {
        panic(err)
    }

    return nil
}

// Set a given tag into the given IFD (tag directory).
func (et *ExifTool) SetTag(ifd string, name string, valueParts []string) (err error) {
    defer func() {
        if r := recover(); r != nil {
            err = r.(error)
        }
    }()

    cmd := make([]string, 0)

    // If they don't want to modify the original file, they can copy it first. 
    // It's either they do it or we do it and we don't want to necessitate 
    // extra operations if not necessary.
    cmd = append(cmd, "--create-exif", "--tag", name, "--ifd", ifd, "-o", et.filepath)

    valuePartsPhrase := strings.Join(valueParts, " ")
    cmd = append(cmd, "--set-value", valuePartsPhrase, et.filepath)

    _, err = et.getBytesFromCommand(ExifToolFilename, cmd)
    if err != nil {
        panic(err)
    }

    return nil
}

// Check if there is a thumbnail available.
func (et *ExifTool) HasThumbnail() (found bool, err error) {
    defer func() {
        if r := recover(); r != nil {
            found = false
            err = r.(error)
        }
    }()

    cmd := make([]string, 0)
    cmd = append(cmd, et.filepath)

    lines, err := et.getLinesFromCommand(ExifToolFilename, cmd)
    if err != nil {
        panic(err)
    }

    o := len(lines) - ThumbnailSearchTailLineCount
    prefixLen := len(ThumbnailExistsLinePrefix)

    for _, line := range lines[o:] {
        if len(line) > prefixLen && 
           line[:prefixLen] == ThumbnailExistsLinePrefix {
            found = true
            break
        }
    }

    return found, nil
}

// Return a thumbnail.
func (et *ExifTool) GetThumbnail() (f *os.File, err error) {
    defer func() {
        if x := recover(); x != nil {
            if f != nil {
                f.Close()
            }

            f = nil
            err = x.(error)
        }
    }()

    cmd := make([]string, 0)

    f, err = ioutil.TempFile("", "thumb")
    if err != nil {
        panic(err)
    }

    cmd = append(cmd, "--extract-thumbnail", "-o", f.Name(), et.filepath)

    _, err = et.getBytesFromCommand(ExifToolFilename, cmd)
    if err != nil {
        panic(err)
    }

    filepath := f.Name()
    f.Close()

    // Now that we've written the file, reopen.
    f, err = os.Open(filepath)
    if err != nil {
        panic(err)
    }

    return f, nil
}
