package streetview

import (
  "fmt"
  "io"
  "os"
  "http"
  "image"
  "image/jpeg"
  "math"
  "json"
  "strconv"
)

type CompositeImage struct {
  // Tiles which comprise image.
  Tiles []image.Image
  // Dimensions of whole image.
  Width, Height int
  // Dimensions of every tile.
  TileWidth, TileHeight int
  // Number of tiles in grid.
  NumX, NumY int
}

func (img *CompositeImage) ColorModel() image.ColorModel {
  // Assume all tiles have the same color model.
  return img.Tiles[0].ColorModel()
}

func (img *CompositeImage) Bounds() image.Rectangle {
  return image.Rect(0, 0, img.Width, img.Height)
}

func (img *CompositeImage) At(x, y int) image.Color {
  x, dx := x / img.TileWidth, x % img.TileWidth
  y, dy := y / img.TileHeight, y % img.TileHeight
  i := x * img.NumY + y

  u, v := float64(dx), float64(dy)
  u = u / float64(img.TileWidth - 1) * float64(img.Tiles[i].Bounds().Dx() - 1)
  v = v / float64(img.TileHeight - 1) * float64(img.Tiles[i].Bounds().Dy() - 1)
  dx, dy = int(u + 0.5), int(v + 0.5)

  return img.Tiles[i].At(dx, dy)
}

// Panorama details which need to be known to download the tiles.
type Details struct {
  ImageWidth, ImageHeight int
  TileWidth, TileHeight int
  PanoId string
}

func (d *Details) MaxZoom() int {
  // Need to fit image in a tile at minimum zoom.
  // Calculate log of ratios in x and y.
  x := math.Log2(float64(d.ImageWidth) / float64(d.TileWidth))
  y := math.Log2(float64(d.ImageHeight) / float64(d.TileHeight))

  // Take maximum.
  return int(math.Ceil(math.Fmax(x, y)))
}

func (d *Details) SizeAtZoom(zoom int) (width, height int) {
  // Divide image dimensions by 2^n, where n is the difference in zoom.
  n := uint(d.MaxZoom() - zoom)
  width = d.ImageWidth >> n
  height = d.ImageHeight >> n
  return width, height
}

func GetDetails(client *http.Client, host, panoId string) (*Details, os.Error) {
  url := detailsUrl(host, panoId, "json")

  // Request the URL.
  response, err := client.Get(url)
  if err != nil {
    return nil, err
  }

  // Parse the response.
  return parseDetails(response.Body)
}

type jsonDetails struct {
  Data struct {
    Image_Width, Image_Height string
    Tile_Width, Tile_Height string
  }
  Location struct {
    PanoId string
  }
}

func parseDetails(body io.Reader) (*Details, os.Error) {
  // Attempt to decode JSON object.
  var raw jsonDetails
  err := json.NewDecoder(body).Decode(&raw)
  if err != nil { return nil, err }

  // Extract details from raw JSON object.
  imageWidth, err := strconv.Atoi(raw.Data.Image_Width)
  if err != nil { return nil, err }
  imageHeight, err := strconv.Atoi(raw.Data.Image_Height)
  if err != nil { return nil, err }
  tileWidth, err := strconv.Atoi(raw.Data.Tile_Width)
  if err != nil { return nil, err }
  tileHeight, err := strconv.Atoi(raw.Data.Tile_Height)
  if err != nil { return nil, err }
  panoId := raw.Location.PanoId

  // Construct details object.
  details := Details{imageWidth, imageHeight, tileWidth, tileHeight, panoId}
  return &details, nil
}

// Parameters:
// host -- e.g. http://cbk0.google.com
// panoId -- Street View panorama ID.
// format -- "json" or "xml"
func detailsUrl(host, panoId, format string) string {
  return fmt.Sprintf("%v/cbk?output=%v&panoid=%v", host, format, panoId)
}

func GetPanorama(details *Details, client *http.Client, host string, zoom int) *CompositeImage {
  // Get size at specified zoom.
  width, height := details.SizeAtZoom(zoom)
  // Determine number of tiles in each direction.
  nx := ceilDivide(width, details.TileWidth)
  ny := ceilDivide(height, details.TileHeight)

  // Make an asynchronous request for each tile.
  chs := make([]<-chan image.Image, nx * ny)
  for x := 0; x < nx; x++ {
    for y := 0; y < ny; y++ {
      i := x * ny + y
      chs[i] = downloadTileAsync(client, host, details.PanoId, zoom, x, y)
    }
  }

  // Synchronize requests and get all tiles.
  tiles := make([]image.Image, nx * ny)
  for i := 0; i < nx * ny; i++ {
    tiles[i] = <-chs[i]
  }

  return &CompositeImage{tiles, width, height, details.TileWidth, details.TileHeight, nx, ny}
}

func downloadTileAsync(client *http.Client, host, panoid string, zoom, x, y int) <-chan image.Image {
  // Create a channel and pass the image back along it.
  ch := make(chan image.Image)
  go func() {
    img, _ := downloadTile(client, host, panoid, zoom, x, y)
    ch <- img
  }()
  return ch
}

func downloadTile(client *http.Client, host, panoid string, zoom, x, y int) (image.Image, os.Error) {
  // Build URL to fetch.
  url := tileUrl(host, panoid, zoom, x, y)

  // Request the URL.
  response, err := client.Get(url)
  if err != nil {
    return nil, err
  }

  // Decode image from response.
  img, err := jpeg.Decode(response.Body)
  if err != nil {
    return nil, err
  }

  return img, nil
}

func tileUrl(host, panoid string, zoom, x, y int) string {
  return fmt.Sprintf("%v/cbk?output=tile&panoid=%v&zoom=%v&x=%v&y=%v", host, panoid, zoom, x, y)
}

func ceilDivide(a, b int) int {
  return (a + b - 1) / b
}
