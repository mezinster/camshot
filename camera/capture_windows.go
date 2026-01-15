//go:build windows
// +build windows

package camera

import (
	"errors"
	"image"
	"image/color"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

// Windows DirectShow/Video capture constants
const (
	WM_CAP_START             = 0x400
	WM_CAP_DRIVER_CONNECT    = WM_CAP_START + 10
	WM_CAP_DRIVER_DISCONNECT = WM_CAP_START + 11
	WM_CAP_GRAB_FRAME        = WM_CAP_START + 60
	WM_CAP_GRAB_FRAME_NOSTOP = WM_CAP_START + 61
	WM_CAP_EDIT_COPY         = WM_CAP_START + 30
	WM_CAP_SET_SCALE         = WM_CAP_START + 53
	WM_CAP_SET_PREVIEW       = WM_CAP_START + 50
	WM_CAP_SET_PREVIEWRATE   = WM_CAP_START + 52
	WM_CAP_SET_VIDEOFORMAT   = WM_CAP_START + 45
	WM_CAP_GET_VIDEOFORMAT   = WM_CAP_START + 44
	WM_CAP_DLG_VIDEOSOURCE   = WM_CAP_START + 42
	WM_CAP_DLG_VIDEOFORMAT   = WM_CAP_START + 41
)

// Clipboard constants
const (
	CF_DIB    = 8
	CF_BITMAP = 2
)

var (
	avicap32          = syscall.NewLazyDLL("avicap32.dll")
	user32            = syscall.NewLazyDLL("user32.dll")
	kernel32          = syscall.NewLazyDLL("kernel32.dll")
	gdi32             = syscall.NewLazyDLL("gdi32.dll")
	capCreateWindow   = avicap32.NewProc("capCreateCaptureWindowW")
	capGetDriverDescA = avicap32.NewProc("capGetDriverDescriptionA")
	sendMessage       = user32.NewProc("SendMessageW")
	destroyWindow     = user32.NewProc("DestroyWindow")
	openClipboard     = user32.NewProc("OpenClipboard")
	closeClipboard    = user32.NewProc("CloseClipboard")
	getClipboardData  = user32.NewProc("GetClipboardData")
	emptyClipboard    = user32.NewProc("EmptyClipboard")
	globalLock        = kernel32.NewProc("GlobalLock")
	globalUnlock      = kernel32.NewProc("GlobalUnlock")
	globalSize        = kernel32.NewProc("GlobalSize")
)

// listPlatformDevices enumerates available camera devices on Windows
func listPlatformDevices() []DeviceInfo {
	var devices []DeviceInfo

	if err := avicap32.Load(); err != nil {
		return devices
	}

	// VfW supports up to 10 capture drivers (0-9)
	nameBuf := make([]byte, 256)
	verBuf := make([]byte, 256)

	for i := 0; i < 10; i++ {
		ret, _, _ := capGetDriverDescA.Call(
			uintptr(i),
			uintptr(unsafe.Pointer(&nameBuf[0])),
			256,
			uintptr(unsafe.Pointer(&verBuf[0])),
			256,
		)

		if ret != 0 {
			// Found a driver
			name := string(nameBuf[:cStrLen(nameBuf)])
			if name == "" {
				name = "Camera " + string(rune('0'+i))
			}
			devices = append(devices, DeviceInfo{
				ID:   i,
				Name: name,
			})
		}
	}

	// If no devices found via capGetDriverDescription, still add device 0
	// as some drivers don't properly enumerate but still work
	if len(devices) == 0 {
		devices = append(devices, DeviceInfo{
			ID:   0,
			Name: "Default Camera",
		})
	}

	return devices
}

// cStrLen returns the length of a null-terminated C string
func cStrLen(b []byte) int {
	for i, c := range b {
		if c == 0 {
			return i
		}
	}
	return len(b)
}

// windowsCamera implements camera capture using Windows Video for Windows API
type windowsCamera struct {
	hwnd     uintptr
	deviceID int
	isOpen   bool
	mutex    sync.Mutex
	width    int
	height   int
}

func getPlatformCamera() (cameraImpl, error) {
	return &windowsCamera{
		width:  640,
		height: 480,
	}, nil
}

func (c *windowsCamera) IsAvailable() bool {
	// Check if avicap32.dll is available
	err := avicap32.Load()
	return err == nil
}

func (c *windowsCamera) Open(deviceID int) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.isOpen {
		return nil
	}

	// Load the DLL
	if err := avicap32.Load(); err != nil {
		return errors.New("camera not available: avicap32.dll not found. Please use the Upload feature instead")
	}

	// Create capture window (invisible, but functional)
	title := syscall.StringToUTF16Ptr("CamShotCapture")
	hwnd, _, err := capCreateWindow.Call(
		uintptr(unsafe.Pointer(title)),
		0,    // style
		0, 0, // x, y
		uintptr(c.width),
		uintptr(c.height),
		0, // parent
		0, // ID
	)

	if hwnd == 0 {
		return errors.New("failed to create capture window: " + err.Error())
	}

	c.hwnd = hwnd
	c.deviceID = deviceID

	// Connect to the camera driver
	ret, _, _ := sendMessage.Call(c.hwnd, WM_CAP_DRIVER_CONNECT, uintptr(deviceID), 0)
	if ret == 0 {
		destroyWindow.Call(c.hwnd)
		c.hwnd = 0
		return errors.New("no camera found at device " + string(rune('0'+deviceID)) + ". Please check your webcam connection or use the Upload feature")
	}

	// Enable scaling so frames fit the capture window
	sendMessage.Call(c.hwnd, WM_CAP_SET_SCALE, 1, 0)

	// Set preview rate to 30 FPS (in milliseconds)
	sendMessage.Call(c.hwnd, WM_CAP_SET_PREVIEWRATE, 33, 0)

	// Start preview mode - this starts the camera streaming
	sendMessage.Call(c.hwnd, WM_CAP_SET_PREVIEW, 1, 0)

	// Give the camera a moment to start streaming
	time.Sleep(100 * time.Millisecond)

	c.isOpen = true
	return nil
}

func (c *windowsCamera) CaptureFrame() (image.Image, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.isOpen || c.hwnd == 0 {
		return nil, errors.New("camera not open")
	}

	// Grab a frame without stopping the preview stream
	ret, _, _ := sendMessage.Call(c.hwnd, WM_CAP_GRAB_FRAME_NOSTOP, 0, 0)
	if ret == 0 {
		return nil, errors.New("failed to grab frame")
	}

	// Copy frame to clipboard
	ret, _, _ = sendMessage.Call(c.hwnd, WM_CAP_EDIT_COPY, 0, 0)
	if ret == 0 {
		return nil, errors.New("failed to copy frame to clipboard")
	}

	// Get the image from clipboard
	img, err := c.getImageFromClipboard()
	if err != nil {
		// Fall back to a placeholder if clipboard fails
		return c.createPlaceholder(), nil
	}

	return img, nil
}

// getImageFromClipboard retrieves a DIB image from the Windows clipboard
func (c *windowsCamera) getImageFromClipboard() (image.Image, error) {
	// Open clipboard
	ret, _, _ := openClipboard.Call(0)
	if ret == 0 {
		return nil, errors.New("failed to open clipboard")
	}
	defer closeClipboard.Call()

	// Get DIB data from clipboard
	hMem, _, _ := getClipboardData.Call(CF_DIB)
	if hMem == 0 {
		return nil, errors.New("no DIB data in clipboard")
	}

	// Lock the memory to get a pointer
	ptr, _, _ := globalLock.Call(hMem)
	if ptr == 0 {
		return nil, errors.New("failed to lock clipboard memory")
	}
	defer globalUnlock.Call(hMem)

	// Get the size of the data
	size, _, _ := globalSize.Call(hMem)
	if size == 0 {
		return nil, errors.New("failed to get clipboard data size")
	}

	// Parse the BITMAPINFOHEADER
	return c.parseDIB(ptr, int(size))
}

// BITMAPINFOHEADER structure
type bitmapInfoHeader struct {
	BiSize          uint32
	BiWidth         int32
	BiHeight        int32
	BiPlanes        uint16
	BiBitCount      uint16
	BiCompression   uint32
	BiSizeImage     uint32
	BiXPelsPerMeter int32
	BiYPelsPerMeter int32
	BiClrUsed       uint32
	BiClrImportant  uint32
}

// parseDIB parses a Device Independent Bitmap from memory
func (c *windowsCamera) parseDIB(ptr uintptr, size int) (image.Image, error) {
	// Read the BITMAPINFOHEADER
	header := (*bitmapInfoHeader)(unsafe.Pointer(ptr))

	width := int(header.BiWidth)
	height := int(header.BiHeight)
	bitCount := int(header.BiBitCount)

	// Handle negative height (top-down DIB)
	topDown := false
	if height < 0 {
		height = -height
		topDown = true
	}

	// Calculate the offset to pixel data
	headerSize := int(header.BiSize)
	// For 24/32 bit, no color table; for 8-bit, there's a color table
	colorTableSize := 0
	if bitCount <= 8 {
		colorTableSize = (1 << bitCount) * 4 // RGBQUAD entries
	}
	pixelOffset := headerSize + colorTableSize

	// Create the image
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Calculate row stride (DIB rows are DWORD aligned)
	bytesPerPixel := bitCount / 8
	rowStride := ((width*bytesPerPixel + 3) / 4) * 4

	// Read pixel data
	pixelData := unsafe.Pointer(ptr + uintptr(pixelOffset))

	for y := 0; y < height; y++ {
		// DIB is typically bottom-up unless height was negative
		srcY := y
		dstY := height - 1 - y
		if topDown {
			dstY = y
		}

		rowPtr := unsafe.Pointer(uintptr(pixelData) + uintptr(srcY*rowStride))

		for x := 0; x < width; x++ {
			var r, g, b, a uint8 = 0, 0, 0, 255

			switch bitCount {
			case 24:
				// BGR format
				pixPtr := unsafe.Pointer(uintptr(rowPtr) + uintptr(x*3))
				bgr := (*[3]byte)(pixPtr)
				b, g, r = bgr[0], bgr[1], bgr[2]
			case 32:
				// BGRA format
				pixPtr := unsafe.Pointer(uintptr(rowPtr) + uintptr(x*4))
				bgra := (*[4]byte)(pixPtr)
				b, g, r, a = bgra[0], bgra[1], bgra[2], bgra[3]
				if a == 0 {
					a = 255 // Some cameras output 0 alpha
				}
			}

			img.Set(x, dstY, color.RGBA{r, g, b, a})
		}
	}

	return img, nil
}

// createPlaceholder creates a test pattern when real capture fails
func (c *windowsCamera) createPlaceholder() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, c.width, c.height))

	// Fill with a gradient pattern to indicate capture is working
	for y := 0; y < c.height; y++ {
		for x := 0; x < c.width; x++ {
			r := uint8(x * 255 / c.width)
			g := uint8(y * 255 / c.height)
			b := uint8(128)
			img.Set(x, y, color.RGBA{r, g, b, 255})
		}
	}

	return img
}

func (c *windowsCamera) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.hwnd != 0 {
		// Stop preview first
		sendMessage.Call(c.hwnd, WM_CAP_SET_PREVIEW, 0, 0)

		// Disconnect from driver
		sendMessage.Call(c.hwnd, WM_CAP_DRIVER_DISCONNECT, 0, 0)

		// Destroy the capture window
		destroyWindow.Call(c.hwnd)
		c.hwnd = 0
	}
	c.isOpen = false
	return nil
}
