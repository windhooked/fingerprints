package main

import (
	"image"
	"image/color"
	"math"
	"sync"
)

var (
	SobelDx = &kernel3x3{size: 3, mat: [3][3]int{{-1, 0, 1}, {-2, 0, 2}, {-1, 0, 1}}}
	SobelDy = &kernel3x3{size: 3, mat: [3][3]int{{-1, -2, -1}, {0, 0, 0}, {1, 2, 1}}}
	Sum9x9  = &sum9x9{size: 9}
)

type Kernel interface {
	Apply(in *image.Gray, x, y int) int
}

type kernel3x3 struct {
	size int
	mat  [3][3]int
}

type sum9x9 struct {
	size int
}

func ApplyKernelAsync(in *image.Gray, out *image.Gray, kernel Kernel) *sync.WaitGroup {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		ApplyKernel(kernel, in, out)
		wg.Done()
	}()
	return wg
}

func ApplyKernel(kernel Kernel, in *image.Gray, out *image.Gray) {
	var min, max int
	min = math.MaxInt64

	for x := 1; x <= in.Bounds().Dx()-1; x++ {
		for y := 1; y <= in.Bounds().Dy()-1; y++ {
			val := kernel.Apply(in, x, y)
			if val > max {
				max = val
			}
			if val < min {
				min = val
			}
		}
	}
	for x := 1; x <= in.Bounds().Dx()-1; x++ {
		for y := 1; y <= in.Bounds().Dy()-1; y++ {
			val := kernel.Apply(in, x, y)
			normVal := uint8(math.MaxUint8 * float64(val-min) / float64(max-min))
			out.SetGray(x, y, color.Gray{Y: normVal})
		}
	}
}

func (k *kernel3x3) Apply(in *image.Gray, x, y int) int {
	sum := 0
	for i := -(k.size - 1) / 2; i <= (k.size-1)/2; i++ {
		for j := -1; j <= 1; j++ {
			sum += k.mat[j+1][i+1] * int(in.GrayAt(x+i, y+j).Y)
		}
	}
	return sum
}

func (k *sum9x9) Apply(in *image.Gray, x, y int) int {
	sum := 0
	for i := -(k.size - 1) / 2; i <= (k.size-1)/2; i++ {
		for j := -(k.size - 1) / 2; j <= (k.size-1)/2; j++ {
			sum += int(in.GrayAt(x+i, y+j).Y)
		}
	}
	return sum
}
