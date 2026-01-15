package camera

import (
	"errors"
	"image"
)

// Camera represents a webcam device
type Camera struct {
	deviceID int
	impl     cameraImpl
}

// cameraImpl is the interface for platform-specific camera implementations
type cameraImpl interface {
	Open(deviceID int) error
	CaptureFrame() (image.Image, error)
	Close() error
	IsAvailable() bool
}

// NewCamera creates a new camera instance for the specified device
func NewCamera(deviceID int) (*Camera, error) {
	c := &Camera{
		deviceID: deviceID,
	}

	// Try to get platform-specific implementation
	impl, err := getPlatformCamera()
	if err != nil {
		return nil, err
	}

	c.impl = impl

	// Try to open the camera
	err = c.impl.Open(deviceID)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// CaptureFrame captures a single frame from the camera
func (c *Camera) CaptureFrame() (image.Image, error) {
	if c.impl == nil {
		return nil, errors.New("camera not initialized")
	}
	return c.impl.CaptureFrame()
}

// Close releases the camera resources
func (c *Camera) Close() error {
	if c.impl == nil {
		return nil
	}
	return c.impl.Close()
}

// IsAvailable checks if camera capture is available on this system
func IsAvailable() bool {
	impl, err := getPlatformCamera()
	if err != nil {
		return false
	}
	return impl.IsAvailable()
}

// DeviceInfo contains information about a camera device
type DeviceInfo struct {
	ID   int
	Name string
}

// ListDevices returns a list of available camera devices
func ListDevices() []DeviceInfo {
	return listPlatformDevices()
}
