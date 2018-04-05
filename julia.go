package main

import (
    "fmt"
    "math/cmplx"
    "strconv"
    "math/rand"
    //"math"
    "time"
    "flag"
    "sync"
    "github.com/gilmae/rescale"
    //"sort"
)

var  maxIterations float64 = 1000.0
var  bailout float64 = 4.0
var  width int = 1600
var height int = 1600
var x int = 0
var y int = 0

type Key struct {
    x, y int
}

var points_map map[Key]Point

const (
  rMin   = -1.5
  rMax   = 1.5
  iMin   = -1.5
  iMax   = 1.5
  usage  = "mandelbot OPTIONS\n\nPlots the mandelbrot set, centered at a point indicated by the provided real and imaginary, and at the given zoom level.\n\nSaves the output into the given path.\n\n"
  default_gradient  = `[["0.0", "000764"],["0.16", "026bcb"],["0.42", "edffff"],["0.6425", "ffaa00"],["0.8675", "000200"],["1.0","000764"]]`
)

func pow(x complex128, y int) complex128 {
  result := x
  for iteration := 0; iteration < y-1; iteration++ {
    result = result * x;
  }
  return result;
}

func calculate_escape(p Point, c complex128, exponent int) Point {
  var iteration float64
  var z complex128 = p.C
  
  for iteration = 0.0;cmplx.Abs(z) < bailout && iteration < maxIterations; iteration+=1 {
    z = pow(z, exponent)+c;
  }

  if (iteration >= maxIterations) {
    return Point{p.C, p.X, p.Y, maxIterations, z, p.ConstantPoint, false}
  }
  
  return Point{p.C, p.X, p.Y, iteration, z, p.ConstantPoint, true}
}

func plot(c complex128, midX float64, midY float64, zoom float64, width int, height int, exponent int, calculated chan Point) {
  points := make(chan Point, 64)

  // spawn four worker goroutines
  var wg sync.WaitGroup
  for i := 0; i < 100; i++ {
    wg.Add(1)
    go func() {
      for p := range points {
        calculated <- calculate_escape(p, c, exponent)
      }
      wg.Done()
    }()
  }

  // Derive new bounds based on focal point and zoom
  new_r_start, new_r_end := rescale.Get_Zoomed_Bounds(rMin, rMax, midX, zoom)
  new_i_start, new_i_end := rescale.Get_Zoomed_Bounds(iMin, iMax, midY, zoom)


  // Pregenerate all the values of the x  & Y CoOrdinates
  xCoOrds := make([]float64, width)
  for i,_ := range xCoOrds {
    xCoOrds[i] = rescale.Rescale(new_r_start, new_r_end, width, i);
  }

  yCoOrds := make([]float64, height)
  for i,_ := range yCoOrds {
    yCoOrds[height-i-1] = rescale.Rescale(new_i_start, new_i_end, height, i);
  }

  for x:=0; x < width; x += 1 {
    for y:=height-1; y >= 0; y -= 1 {
      points <- Point{complex(xCoOrds[x], yCoOrds[y]),x,y, 0, complex(0,0), c, false}
    }
  }

  close(points)

  wg.Wait()
}

func get_cordinates(midX float64, midY float64, zoom float64, width int, height int, x int, y int) complex128 {
  new_r_start, new_r_end := rescale.Get_Zoomed_Bounds(rMin, rMax, midX, zoom)
  scaled_r := rescale.Rescale(new_r_start, new_r_end, width, x)

  new_i_start, new_i_end := rescale.Get_Zoomed_Bounds(iMin, iMax, midY, zoom)
  scaled_i := rescale.Rescale(new_i_start, new_i_end, height, height-y)

  return complex(scaled_r, scaled_i)
}
    
func main() {
  //start := time.Now()

  var midX float64
  var midY float64
  var cr float64
  var ci float64
  var zoom float64
  var output string
  var filename string
  var gradient string
  var exponent int
  var mode string

  rand.Seed(time.Now().UnixNano())
  flag.Float64Var(&midX, "r", 0.0, "Real component of the midpoint.")
  flag.Float64Var(&midY, "i", 0.0, "Imaginary component of the midpoint.")
  flag.Float64Var(&cr, "cr", 0.0, "Real component of the c constant.")
  flag.Float64Var(&ci, "ci", 0.0, "Imaginary component of the c constant.")
  flag.IntVar(&exponent, "e", 2, "The exponent to raise z to.")
  flag.Float64Var(&zoom, "z", 1, "Zoom level.")
  flag.StringVar(&output, "o", ".", "Output path.")
  flag.StringVar(&filename, "f", "", "Output file name.")
  flag.StringVar(&colour_mode, "c", "none", "Colour mode: true, smooth, banded, none.")
  flag.Float64Var(&bailout, "b", 4.0, "Bailout value.")
  flag.IntVar(&width, "w", 1600, "Width of render.")
  flag.IntVar(&height, "h", 1600, "Height of render.")
  flag.Float64Var(&maxIterations, "m", 2000.0, "Maximum Iterations before giving up on finding an escape.")
  flag.StringVar(&gradient, "g", default_gradient, "Gradient to use.")
  flag.StringVar(&mode, "mode", "image", "Mode: image, coordsAt")
  flag.IntVar(&x, "x", 0, "x cordinate of a pixel, used for translating to the real component. 0,0 is top left.")
  flag.IntVar(&y, "y", 0, "y cordinate of a pixel, used for translating to the real component. 0,0 is top left.")
  flag.Parse()

  points_map = make(map[Key]Point)

  calculatedChan := make(chan Point)

  go func(points<-chan Point, hash map[Key]Point) {
    for p := range points {
      hash[Key{p.X,p.Y}] = p
    }
  }(calculatedChan, points_map)

  
  if (mode == "image") {
    plot(complex(cr,ci),midX, midY, zoom, width, height, exponent, calculatedChan)
    if (filename == "") {
      filename = "julia_c_" +strconv.FormatFloat(cr, 'E', -1, 64) + "+ " + strconv.FormatFloat(ci, 'E', -1, 64) + "i_" + strconv.FormatFloat(midX, 'E', -1, 64) + "_" + strconv.FormatFloat(midY, 'E', -1, 64) + "_" +  strconv.FormatFloat(zoom, 'E', -1, 64) + ".jpg"
    }

    filename = output + "/" + filename

    draw_image(filename, points_map, width, height, gradient, mode=="image" && colour_mode=="smooth")
    fmt.Printf("%s\n", filename)
  } else if (mode == "coordsAt") {
    
    var p = get_cordinates(midX, midY, zoom, width, height, x, y)
    fmt.Printf("%18.17e, %18.17e\n", real(p), imag(p))
  }
}
