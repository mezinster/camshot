package ui

import (
	"camshot/camera"
	"image"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// CameraView handles camera capture functionality
type CameraView struct {
	app            *App
	container      *fyne.Container
	previewImage   *canvas.Image
	statusLabel    *widget.Label
	capturedImage  image.Image
	cam            *camera.Camera
	isStreaming    bool
	stopChan       chan bool
	selectedDevice int
	deviceSelect   *widget.Select
}

// NewCameraView creates a new camera view
func NewCameraView(app *App) *CameraView {
	v := &CameraView{
		app:      app,
		stopChan: make(chan bool, 1), // Buffered to prevent deadlock
	}
	v.setup()
	return v
}

// setup initializes the camera view UI
func (v *CameraView) setup() {
	// Create placeholder for camera preview
	v.previewImage = canvas.NewImageFromImage(nil)
	v.previewImage.FillMode = canvas.ImageFillContain
	v.previewImage.SetMinSize(fyne.NewSize(640, 480))

	// Status label
	v.statusLabel = widget.NewLabel("Camera not started")
	v.statusLabel.Alignment = fyne.TextAlignCenter

	// Camera device selector
	devices := camera.ListDevices()
	deviceNames := make([]string, len(devices))
	deviceIDs := make(map[string]int)
	for i, dev := range devices {
		deviceNames[i] = dev.Name
		deviceIDs[dev.Name] = dev.ID
	}

	if len(deviceNames) == 0 {
		deviceNames = []string{"No cameras found"}
	}

	v.deviceSelect = widget.NewSelect(deviceNames, func(selected string) {
		if id, ok := deviceIDs[selected]; ok {
			v.selectedDevice = id
		}
	})
	if len(deviceNames) > 0 {
		v.deviceSelect.SetSelectedIndex(0)
		if len(devices) > 0 {
			v.selectedDevice = devices[0].ID
		}
	}

	// Control buttons
	startBtn := widget.NewButton("Start Camera", func() {
		v.startCamera()
	})

	stopBtn := widget.NewButton("Stop Camera", func() {
		v.stopCamera()
	})

	captureBtn := widget.NewButton("Capture Photo", func() {
		v.capturePhoto()
	})
	captureBtn.Importance = widget.HighImportance

	useBtn := widget.NewButton("Use This Photo", func() {
		v.usePhoto()
	})

	// Refresh cameras button
	refreshBtn := widget.NewButton("Refresh", func() {
		v.refreshDevices()
	})

	// Instructions
	instructions := widget.NewLabel(
		"Select your camera and click 'Start Camera' to begin.\n" +
			"Position your subject and click 'Capture Photo'.",
	)
	instructions.Wrapping = fyne.TextWrapWord
	instructions.Alignment = fyne.TextAlignCenter

	// Device selection row
	deviceRow := container.NewBorder(
		nil, nil,
		widget.NewLabel("Camera:"), refreshBtn,
		v.deviceSelect,
	)

	// Preview container
	previewContainer := container.NewBorder(
		widget.NewLabel("Camera Preview"),
		v.statusLabel,
		nil, nil,
		container.NewCenter(v.previewImage),
	)

	// Buttons row
	buttonsRow := container.NewHBox(
		startBtn,
		stopBtn,
		captureBtn,
		useBtn,
	)

	// Top section with instructions and device selector
	topSection := container.NewVBox(
		instructions,
		deviceRow,
	)

	// Main layout
	v.container = container.NewBorder(
		topSection,
		container.NewCenter(buttonsRow),
		nil, nil,
		previewContainer,
	)
}

// refreshDevices updates the camera device list
func (v *CameraView) refreshDevices() {
	devices := camera.ListDevices()
	deviceNames := make([]string, len(devices))
	for i, dev := range devices {
		deviceNames[i] = dev.Name
	}

	if len(deviceNames) == 0 {
		deviceNames = []string{"No cameras found"}
	}

	v.deviceSelect.Options = deviceNames
	if len(deviceNames) > 0 {
		v.deviceSelect.SetSelectedIndex(0)
		if len(devices) > 0 {
			v.selectedDevice = devices[0].ID
		}
	}
	v.deviceSelect.Refresh()
}

// startCamera initializes and starts the camera
func (v *CameraView) startCamera() {
	if v.isStreaming {
		return
	}

	v.statusLabel.SetText("Starting camera...")

	// Disable device selector while streaming
	v.deviceSelect.Disable()

	// Initialize camera with selected device
	var err error
	v.cam, err = camera.NewCamera(v.selectedDevice)
	if err != nil {
		v.statusLabel.SetText("Camera error: " + err.Error())
		v.deviceSelect.Enable()
		showError(v.app.GetWindow(),
			"Could not access camera. Please ensure:\n"+
				"1. A webcam is connected\n"+
				"2. No other application is using the camera\n"+
				"3. Camera permissions are granted\n\n"+
				"You can still use the 'Upload Photo' feature.")
		return
	}

	v.isStreaming = true
	v.statusLabel.SetText("Camera active - streaming")

	// Start streaming in background
	go v.streamLoop()
}

// streamLoop continuously captures frames from the camera
func (v *CameraView) streamLoop() {
	ticker := time.NewTicker(33 * time.Millisecond) // ~30 FPS
	defer ticker.Stop()

	for {
		select {
		case <-v.stopChan:
			return
		case <-ticker.C:
			if v.cam != nil && v.isStreaming {
				frame, err := v.cam.CaptureFrame()
				if err == nil && frame != nil {
					v.previewImage.Image = frame
					v.previewImage.Refresh()
				}
			}
		}
	}
}

// stopCamera stops the camera stream
func (v *CameraView) stopCamera() {
	if !v.isStreaming {
		return
	}

	v.isStreaming = false

	// Non-blocking send to stop channel
	select {
	case v.stopChan <- true:
	default:
		// Channel already has a value, that's fine
	}

	if v.cam != nil {
		v.cam.Close()
		v.cam = nil
	}

	// Re-enable device selector
	v.deviceSelect.Enable()
	v.statusLabel.SetText("Camera stopped")
}

// capturePhoto captures a single frame
func (v *CameraView) capturePhoto() {
	if v.cam == nil {
		showError(v.app.GetWindow(), "Camera not started. Please start the camera first.")
		return
	}

	frame, err := v.cam.CaptureFrame()
	if err != nil {
		showError(v.app.GetWindow(), "Error capturing photo: "+err.Error())
		return
	}

	v.capturedImage = frame
	v.previewImage.Image = frame
	v.previewImage.Refresh()
	v.statusLabel.SetText("Photo captured! Click 'Use This Photo' to proceed.")
}

// usePhoto sets the captured photo as the current working image
func (v *CameraView) usePhoto() {
	if v.capturedImage == nil {
		showError(v.app.GetWindow(), "No photo captured. Please capture a photo first.")
		return
	}

	v.app.SetImage(v.capturedImage)
	showInfo(v.app.GetWindow(), "Photo loaded! Go to 'Puzzle Settings' tab to configure and generate your puzzle.")
}

// Container returns the view's container
func (v *CameraView) Container() fyne.CanvasObject {
	return v.container
}
