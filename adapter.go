package exiftool

import (
    "fmt"
    "bytes"
    "strings"

    "os/exec"

    "github.com/dsoprea/go-xmlvisitor"
)

const (
    ExifToolFilename = "exif"
)

type ExifVisitor interface {
    HandleTag(tagName *string, value *string) error
}

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
}

func NewExifTool(filepath *string) *ExifTool {
    return &ExifTool {
            filepath: *filepath,
    }
}

func (et *ExifTool) callCommand(cmd string, args []string) (buffer *bytes.Buffer, err error) {
    defer func() {
        if r := recover(); r != nil {
            buffer = nil
            err = r.(error)
        }
    }()

    // It's necessary to pass in at least one non-variadic argument (otherwise 
    // you'll get warned about there not being enough arguments).
    c := exec.Command(cmd, args...)
    
    var output bytes.Buffer
    c.Stdout = &output
    c.Stderr = &output

    err = c.Run()
    if err != nil {
        fmt.Printf(output.String())
        panic(err)
    }

    return &output, nil
}

func (et *ExifTool) ReadTags(ev ExifVisitor) (err error) {
    defer func() {
        if r := recover(); r != nil {
            err = r.(error)
        }
    }()

    cmd := []string { "-x", et.filepath }
    outputXml, err := et.callCommand(ExifToolFilename, cmd)
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

func (et *ExifTool) SetTag(ifd string, name string, valueParts []string, outputFilename string) (err error) {
    cmd := make([]string, 0)

    if outputFilename != "" {
        cmd = append(cmd, "--output", outputFilename)
    }

    cmd = append(cmd, "--create-exif", "--tag", name, "--ifd", ifd)

    valuePartsPhrase := strings.Join(valueParts, " ")
    cmd = append(cmd, "--set-value", valuePartsPhrase, et.filepath)

    _, err = et.callCommand(ExifToolFilename, cmd)
    if err != nil {
        panic(err)
    }

    return nil
}
