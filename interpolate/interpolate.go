package interpolate

import (
  "image"
)

type Image interface {
  At(x, y float64) image.Color
}

type Nearest struct {
  Image image.Image
}

func (self *Nearest) At(x, y float64) image.Color {
  u, v := int(x + 0.5), int(y + 0.5)
  return self.Image.At(u, v)
}

type Bilinear struct {
  Image image.Image
}

func (self *Bilinear) At(x, y float64) image.Color {
  x1, x2 := int(x), int(x + 1)
  y1, y2 := int(y), int(y + 1)
  dx, dy := x - float64(x1), y - float64(y1)

  bottomLeft := self.Image.At(x1, y1)
  bottomRight := self.Image.At(x2, y1)
  topLeft := self.Image.At(x1, y2)
  topRight := self.Image.At(x2, y2)

  top := Mix(topLeft, topRight, dx)
  bottom := Mix(bottomLeft, bottomRight, dx)
  return Mix(bottom, top, dy)
}

func Mix(x, y image.Color, theta float64) image.Color {
  rx, gx, bx, ax := x.RGBA()
  ry, gy, by, ay := y.RGBA()

  r := (1 - theta) * float64(rx) + theta * float64(ry)
  g := (1 - theta) * float64(gx) + theta * float64(gy)
  b := (1 - theta) * float64(bx) + theta * float64(by)
  a := (1 - theta) * float64(ax) + theta * float64(ay)

  return &image.RGBA64Color{uint16(r), uint16(g), uint16(b), uint16(a)}
}
