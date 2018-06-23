package thumbnailer

// #cgo pkg-config: GraphicsMagick++
// #cgo CFLAGS: -std=c11 -O3 -D_POSIX_C_SOURCE -I${SRCDIR}/libimagequant -I${SRCDIR}/lodepng
// #cgo CXXFLAGS: -std=c++17 -O3 -I${SRCDIR}/libimagequant -I${SRCDIR}/lodepng
// #cgo LDFLAGS: -lm
// #include "blur.c"
// #include "kmeans.c"
// #include "libimagequant.c"
// #include "mediancut.c"
// #include "mempool.c"
// #include "nearest.c"
// #include "pam.c"
// #include "init.h"
// #include "thumbnailer.h"
// #include <stdlib.h>
import "C"
import (
	"errors"
	"unsafe"
)

// Image stores an image of known dimensions.
// To decrease allocations call ReturnBuffer() on Data, after you are done using
// Image.
type Image struct {
	Data []byte
	Dims
}

func init() {
	C.magickInit()
}

// processImage generates a thumbnail from a source image buffer. If width and
// height are non-zero, buf is assumed to be a raw RGBA image.
func processImage(src Source, opts Options) (Source, Thumbnail, error) {
	srcC := C.struct_Buffer{
		data:   (*C.uint8_t)(C.CBytes(src.Data)),
		size:   C.size_t(len(src.Data)),
		width:  C.ulong(src.Width),
		height: C.ulong(src.Height),
	}

	optsC := C.struct_Options{
		JPEGCompression: C.uint8_t(opts.JPEGQuality),
		PNGCompression: C.struct_CompressionRange{
			min: C.uint8_t(opts.PNGQuality.Min),
			max: C.uint8_t(opts.PNGQuality.Max),
		},
		maxSrcDims: C.struct_Dims{
			width:  C.ulong(opts.MaxSourceDims.Width),
			height: C.ulong(opts.MaxSourceDims.Height),
		},
		thumbDims: C.struct_Dims{
			width:  C.ulong(opts.ThumbDims.Width),
			height: C.ulong(opts.ThumbDims.Height),
		},
	}

	var thumb C.struct_Thumbnail
	errC := C.thumbnail(&srcC, &thumb, optsC)
	defer func() {
		if thumb.img.data != nil {
			C.free(unsafe.Pointer(thumb.img.data))
		}
		if errC != nil {
			C.free(unsafe.Pointer(errC))
		}
	}()
	if errC != nil {
		var err error
		switch s := C.GoString(errC); s {
		case "too wide":
			err = ErrTooWide
		case "too tall":
			err = ErrTooTall
		default:
			err = errors.New(s)
		}
		return src, Thumbnail{}, err
	}

	src.Width = uint(srcC.width)
	src.Height = uint(srcC.height)
	thumbnail := Thumbnail{
		IsPNG: bool(thumb.isPNG),
		Image: Image{
			Data: copyCBuffer(
				unsafe.Pointer(thumb.img.data),
				C.int(thumb.img.size),
			),
			Dims: Dims{
				Width:  uint(thumb.img.width),
				Height: uint(thumb.img.height),
			},
		},
	}
	return src, thumbnail, nil
}
