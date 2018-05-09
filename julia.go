package main

import (
    "fmt"
    "strconv"
    "math"
    "math/rand"
    "math/cmplx"
    "time"
    "flag"
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



func get_cordinates(midX float64, midY float64, zoom float64, width int, height int, x int, y int) complex128 {
  new_r_start, new_r_end := rescale.GetZoomedBounds(rMin, rMax, midX, zoom)
  scaled_r := rescale.Rescale(new_r_start, new_r_end, width, x)

  new_i_start, new_i_end := rescale.GetZoomedBounds(iMin, iMax, midY, zoom)
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

  var calculator EscapeCalculator = func(z complex128) (float64, complex128, bool) {
      var iteration float64
      var c complex128 = complex(cr,ci)
      
      for iteration = 0.0;cmplx.Abs(z) < bailout && iteration < maxIterations; iteration+=1 {
        z = pow(z, exponent)+c;
      }
      
      if (iteration >= maxIterations) {
        return maxIterations, z, false
      }

  

      z = z*z+c
      z = z*z+c
      iteration += 2
      reZ := real(z)
      imZ := imag(z)
      magnitude := math.Sqrt(reZ * reZ + imZ * imZ)
      mu := iteration + 1 - (math.Log(math.Log(magnitude)))/math.Log(2.0)
      
      return mu, z, true
  }

  var points_map = escape_time_calculator(midX, midY, zoom, width, height, calculator);
  
  if (mode == "image") {
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
