//go:build !windows
// +build !windows

package camera

import (
	"errors"
	"image"
)

// stubCamera is a placeholder for non-Windows systems
type stubCamera struct{}

func getPlatformCamera() (cameraImpl, error) {
	return &stubCamera{}, nil
}

func (c *stubCamera) IsAvailable() bool {
	return false
}

func (c *stubCamera) Open(deviceID int) error {
	return errors.New("camera capture is only available on Windows. Please use the Upload feature to load images")
}

func (c *stubCamera) CaptureFrame() (image.Image, error) {
	return nil, errors.New("camera not available on this platform")
}

func (c *stubCamera) Close() error {
	return nil
}

// listPlatformDevices returns an empty list on non-Windows platforms
func listPlatformDevices() []DeviceInfo {
	return []DeviceInfo{}
}
