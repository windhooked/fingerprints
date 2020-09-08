package main

import (
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"os"
	_ "time"

	"github.com/google/gxui"
	"github.com/google/gxui/drivers/gl"
	"github.com/google/gxui/math"
	"github.com/google/gxui/samples/flags"
	"github.com/nfnt/resize"
	_ "github.com/nfnt/resize"
	fp "github.com/windhooked/fingerprints/api"
)

func loadImage(name string) *image.Gray {
	f, err := os.Open(name)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	m, _, err := image.Decode(f)
	if err != nil {
		panic(err)
	}

	m = resize.Resize(400, 400, m, resize.Bilinear)

	bounds := m.Bounds()
	w, h := bounds.Max.X, bounds.Max.Y
	gray := image.NewGray(bounds)
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			oldColor := m.At(x, y)
			grayColor := color.GrayModel.Convert(oldColor)
			gray.Set(x, y, grayColor)
		}
	}

	return gray
}

func appMain(driver gxui.Driver) {
	original := loadImage("corpus/nist2.jpg")
	img := fp.NewMatrixFromGray(original)
	processImage(driver, img)
}

var posX, posY = 0, 0

func showImage(driver gxui.Driver, title string, in *fp.Matrix) {
	bounds := in.Bounds()

	theme := flags.CreateTheme(driver)
	window := theme.CreateWindow(bounds.Dx(), bounds.Dy(), title)
	window.SetPosition(math.NewPoint(posX, posY))
	posX += bounds.Dx()
	if posX%(4*bounds.Dx()) == 0 {
		posY += bounds.Dy() + 70
		posX = 0
	}
	window.SetScale(flags.DefaultScaleFactor)
	window.SetBackgroundBrush(gxui.WhiteBrush)

	img := theme.CreateImage()
	window.AddChild(img)

	gray := image.NewRGBA(in.Bounds())
	draw.Draw(gray, in.Bounds(), in.ToGray(), image.ZP, draw.Src)
	texture := driver.CreateTexture(gray, 1)
	img.SetTexture(texture)
	window.OnClose(driver.Terminate)
}

func processImage(driver gxui.Driver, in *fp.Matrix) {
	bounds := in.Bounds()
	normalized := fp.NewMatrix(bounds)

	//showImage(driver, "Original", in)
	fp.Normalize(in, normalized)
	showImage(driver, "Normalized", normalized)

	gx, gy := fp.NewMatrix(bounds), fp.NewMatrix(bounds)
	c1 := fp.ParallelConvolution(fp.SobelDx, normalized, gx)
	c2 := fp.ParallelConvolution(fp.SobelDy, normalized, gy)
	c1.Wait()
	c2.Wait()

	//Consistency matrix
	consistency, normConsistency := fp.NewMatrix(bounds), fp.NewMatrix(bounds)
	c1 = fp.ParallelConvolution(fp.NewSqrtKernel(gx, gy), in, consistency)
	c1.Wait()
	fp.Normalize(consistency, normConsistency)
	showImage(driver, "Normalized Consistency", normConsistency)

	// Compute directional
	directional, normDirectional := fp.NewMatrix(bounds), fp.NewMatrix(bounds)
	c1 = fp.ParallelConvolution(fp.NewDirectionalKernel(gx, gy), directional, directional)
	c1.Wait()
	fp.Normalize(directional, normDirectional)
	showImage(driver, "Directional", normDirectional)

	// Compute filtered directional
	filteredD, normFilteredD := fp.NewMatrix(bounds), fp.NewMatrix(bounds)
	fp.Convolute(fp.NewFilteredDirectional(gx, gy, 4), filteredD, filteredD)
	fp.Normalize(filteredD, normFilteredD)
	showImage(driver, "Filtered Directional", normFilteredD)

	// Compute segmented image
	segmented, normSegmented := fp.NewMatrix(bounds), fp.NewMatrix(bounds)
	fp.Convolute(fp.NewVarianceKernel(filteredD, 8), normalized, segmented)
	fp.Normalize(segmented, normSegmented)
	showImage(driver, "Filtered Directional Std Dev.", normSegmented)

	// Compute binarized segmented image
	binarizedSegmented := fp.NewMatrix(bounds)
	fp.Binarize(normSegmented, binarizedSegmented)
	showImage(driver, "Binarized Segmented", binarizedSegmented)

	// Binarize normalized image
	binarizedNorm := fp.NewMatrix(bounds)
	fp.Binarize(normalized, binarizedNorm)
	showImage(driver, "Binarized Normalized", binarizedNorm)

	fp.BinarizeEnhancement(binarizedNorm)
	showImage(driver, "Binarized Enhanced", binarizedNorm)

	// Skeletonize
	fp.Skeletonize(binarizedNorm)
	showImage(driver, "Skeletonized", binarizedNorm)
}

func main() {
	gl.StartDriver(appMain)
}
