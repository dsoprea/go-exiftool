## Description

This library allows you to read and write EXIF tags and to extract a thumbnail for an image.

## Dependencies

- Go 1.5
- The "exif" tool (*exif* via Brew on Mac OS X or Apt on Ubuntu).

## Example

Excerpt (see [exiftooltest](exiftooltest/main.go) for the whole thing):

```go
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

    err = et.SetTag("GPS", "GPSLongitude", []string { "80", "1", "6", "1", "6", "1" })
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
```

Output:

```
Dumping tags.

TAG [Image_Width]=[4128]
TAG [Image_Length]=[2322]
TAG [Manufacturer]=[samsung]
TAG [Model]=[SM-N900V]
TAG [Orientation]=[Right-top]
TAG [X-Resolution]=[72]
TAG [Y-Resolution]=[72]
TAG [Resolution_Unit]=[Inch]
TAG [Software]=[N900VVRUEOF1]
TAG [Date_and_Time]=[2016:01:18 20:36:22]
TAG [YCbCr_Positioning]=[Centered]
TAG [Compression]=[JPEG compression]
TAG [X-Resolution]=[72]
TAG [Y-Resolution]=[72]
TAG [Resolution_Unit]=[Inch]
TAG [F-Number]=[f/2.2]
TAG [Exposure_Program]=[Normal program]
TAG [Exif_Version]=[Exif Version 2.2]
TAG [Date_and_Time__Original_]=[2016:01:18 20:36:22]
TAG [Date_and_Time__Digitized_]=[2016:01:18 20:36:22]
TAG [Maximum_Aperture_Value]=[2.28 EV (f/2.2)]
TAG [Metering_Mode]=[Center-weighted average]
TAG [Focal_Length]=[4.1 mm]
TAG [Color_Space]=[sRGB]
TAG [Pixel_X_Dimension]=[4128]
TAG [Pixel_Y_Dimension]=[2322]
TAG [Exposure_Mode]=[Auto exposure]
TAG [White_Balance]=[Auto white balance]
TAG [Focal_Length_in_35mm_Film]=[31]
TAG [Scene_Capture_Type]=[Night scene]
TAG [Image_Unique_ID]=[837d96106ec1744c0000000000000000]
TAG [FlashPixVersion]=[FlashPix Version 1.0]
TAG [GPS_Tag_Version]=[2.2.0.0]
TAG [North_or_South_Latitude]=[N]
TAG [Latitude]=[26, 33, 44]
TAG [East_or_West_Longitude]=[W]
TAG [Longitude]=[80,  6,  6]
TAG [Altitude_Reference]=[Sea level]

Setting tag.
Wrote thumbnail: [thumb.jpg]
```

The arguments that are passed to `SetTag()` depend on the type of value that the tag is supposed to have. If *libexif* doesn't get what it needs, execution will terminate with a [fairly helpful] error. 

For reference, this image is geotagged (several tags describing the geographical location where the picture was taken). We update the `Longitude` tag (which is actually a nice alias for the standard `GPSLongitude` name) which takes an unsigned rational (read: "set of fractions") value. Usually this expects a principal longitude (with a denominator of 1) followed by longitudinal minutes and seconds (both having a denominator of 60). However, here we just want to write back the values that were there to begin with. So we just pass { "80", "1", "6", "1", "6", "1" } so that we keep what we had and no division is performed.

## Implementation Notes

- The original plan was to directly integrate the *libexif* library but it became a non-trivial effort to find the correct libjpeg flavor and version. Then, I started parsing JPEG files (which were the primary use-case, though other image types support EXIF) directly in order to extract and update the EXIF data. However, even though the JPEG file is structured into individually indexed pieces, you can't successfully parse the entire file without being required to decode the image data (most segments, except for the actual image data, have length prefixes). As this library was merely a stepping-stone in a larger effort and the *exif* command-line tool is available for most environment a decision was made to just call that.

- For convenience, `SetTag()` always passes the flag to create an EXIF segment if it doesn't already exist.

- We won't add an EXIF segment to an image that is missing one if we're just reading tags. So, use the error return to make an assumption about whether one exists.

## TODO

- Add ability to remove a tag or IFD (tag directory).
- Add ability to remove a thumbnail.
- Add ability to set a thumbnail.
