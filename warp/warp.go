package warp

import (
  "image"
  "interpolate"
)

type Map func(int, int) (float64, float64)

func Warp(src interpolate.Image, width, height int, f Map) image.Image {
  out := image.NewRGBA(width, height)

  for x := 0; x < width; x++ {
    for y := 0; y < height; y++ {
      u, v := f(x, y)
      out.Set(x, y, src.At(u, v))
    }
  }

  return out
}
