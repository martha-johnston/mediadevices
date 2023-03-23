package camera

import (
	"fmt"
	"image"

	"github.com/pion/mediadevices/pkg/avfoundation"
	"github.com/pion/mediadevices/pkg/driver"
	"github.com/pion/mediadevices/pkg/frame"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
)

type camera struct {
	device  avfoundation.Device
	session *avfoundation.Session
	rcClose func()
}

func init() {
	Initialize()
}

// Initialize finds and registers camera devices. This is part of an experimental API.
func Initialize() {
	devices, err := avfoundation.Devices(avfoundation.Video)
	if err != nil {
		panic(err)
	}

	for _, device := range devices {
		cam := newCamera(device)
		driver.GetManager().Register(cam, driver.Info{
			Label:      device.UID,
			DeviceType: driver.Camera,
			Name:       device.Name,
			Connected:  device.Connected,
		})
	}
}

func Connected(oldLabel string) bool {
	newDevices, err := avfoundation.Devices(avfoundation.Video)
	fmt.Print("new:", newDevices)
	if err != nil {
		panic(err)
	}

	for i := 0; i < len(newDevices); i++ {
		if newDevices[i].UID == oldLabel {
			newDevices = nil
			return true
		}
	}
	newDevices = nil
	return false
}

func newCamera(device avfoundation.Device) *camera {
	return &camera{
		device: device,
	}
}

func (cam *camera) Open() error {
	var err error
	cam.session, err = avfoundation.NewSession(cam.device)
	return err
}

func (cam *camera) Close() error {
	if cam.rcClose != nil {
		cam.rcClose()
	}
	val := cam.session.Close()
	return val
}

func (cam *camera) VideoRecord(property prop.Media) (video.Reader, error) {
	decoder, err := frame.NewDecoder(property.FrameFormat)
	if err != nil {
		return nil, err
	}

	rc, err := cam.session.Open(property)
	if err != nil {
		return nil, err
	}
	cam.rcClose = rc.Close
	r := video.ReaderFunc(func() (image.Image, func(), error) {
		frame, _, err := rc.Read()
		if err != nil {
			return nil, func() {}, err
		}
		return decoder.Decode(frame, property.Width, property.Height)
	})
	return r, nil
}

func (cam *camera) Properties() []prop.Media {
	return cam.session.Properties()
}
