package server

import (
  "http"
  "io"
  "bufio"
  "os"
  "streetview"
  "image"
  "image/jpeg"
  "interpolate"
  "warp"
  "math"
  "appengine"
  "appengine/datastore"
  "appengine/blobstore"
  //"appengine/urlfetch"
  "time"
  "fmt"
)

const HOST = "http://cbk0.google.com"

func init() {
  http.HandleFunc("/", showHome)
  http.HandleFunc("/create", createPlanet)
}

type Planet struct {
  PanoId string
  Time datastore.Time
  appengine.BlobKey
}

func showHome(w http.ResponseWriter, r *http.Request) {
  f, err := os.Open("client/home.html")
  if err != nil {
    CriticalError(w, err)
    return
  }

  io.Copy(w, f)
}

func CriticalError(w http.ResponseWriter, err os.Error) {
  http.Error(w, err.String(), http.StatusInternalServerError)
}

func createPlanet(w http.ResponseWriter, r *http.Request) {
  panoId := r.FormValue("panoid")

  if len(panoId) == 0 {
    // Panorama not specified.
    CriticalError(w, os.NewError("panoid not specified"))
    return
  }

  context := appengine.NewContext(r)
  go work(panoId, context)

  fmt.Fprintln(w, "patience is a virtue")
}

func work(panoId string, context appengine.Context) {
  //client := urlfetch.Client(context)
  client := new(http.Client)

  fmt.Println("getting panorama details...")
  // Fetch panorama details.
  details, err := streetview.GetDetails(client, HOST, panoId)
  if err != nil {
    // Could not obtain details on panorama.
    fmt.Println("FAILED")
    return
  }

  // Download the panorama.
  fmt.Println("getting panorama tiles...")
  pano := streetview.GetPanorama(details, client, HOST, 2)

  dstWidth, dstHeight := 800, 800
  srcWidth, srcHeight := pano.Bounds().Dx(), pano.Bounds().Dy()
  f := LogPolarMap(dstWidth, dstHeight, srcWidth, srcHeight, 2)
  in := interpolate.Bilinear{&image.Tiled{pano, image.ZP}}
  fmt.Println("warping image...")
  out := warp.Warp(&in, dstWidth, dstHeight, f)

  fmt.Println("creating blobstore entry...")
  w, err := blobstore.Create(context, "application/octet-stream")
  fmt.Println("FINISHED")
  if err != nil {
    fmt.Println("FAILED")
    return
  }

  b := bufio.NewWriter(w)

  fmt.Println("saving to blobstore...")
  err = jpeg.Encode(b, out, &jpeg.Options{Quality: 75})
  if err != nil {
    fmt.Println("FAILED")
    return
  }

  fmt.Println("saving to datastore...")
  blobKey, err := w.Key()
  planet := Planet{panoId, datastore.SecondsToTime(time.Seconds()), blobKey}
  dataKey, err := datastore.Put(context, datastore.NewIncompleteKey("planet"), &planet)
  if err != nil {
    fmt.Println("FAILED")
    return
  }

  fmt.Println("saved", dataKey)
}

func LogPolarMap(dstWidth, dstHeight, srcWidth, srcHeight int, zoom float64) warp.Map {
  m, n := float64(dstWidth), float64(dstHeight)
  w, h := float64(srcWidth), float64(srcHeight)

  // Calculate maximum radius and scale factor.
  R := math.Sqrt(m * m + n * n) / 2 * zoom
  A := R / (math.Exp((2 * math.Pi * h) / w) - 1)

  return func(x, y int) (float64, float64) {
    // Co-ordinates relative to center.
    i := float64(x) - (m - 1) / 2
    j := float64(y) - (n - 1) / 2

    // Get log-polar co-ordinates.
    r := math.Sqrt(i * i + j * j)
    theta := math.Atan2(i, j)
    if theta < 0 {
      theta += 2 * math.Pi
    }

    // Map to log-polar co-ordinates.
    p := w / (2 * math.Pi) * math.Log(r / A + 1)
    q := w / (2 * math.Pi) * theta
    // Invert coordinates.
    p = (h - 1) - p
    q = (w - 1) - q

    return q, p
  }
}
