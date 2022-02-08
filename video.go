package main

import (
	"bytes"
	"image"
	"image/jpeg"
	"log"
	"time"

	"github.com/blackjack/webcam"
	"github.com/saljam/mjpeg"
)

// V4L format identifiers from /usr/include/linux/videodev2.h.
const (
	MJPEG   webcam.PixelFormat = 1196444237
	YUYV422 webcam.PixelFormat = 1448695129
)

// Width, Height.
var CamResolutions = map[int][]int{
	1:  {160, 120},
	2:  {176, 144},
	3:  {320, 176},
	4:  {320, 240},
	5:  {352, 288},
	6:  {432, 240},
	7:  {544, 288},
	8:  {640, 360},
	9:  {640, 480},
	10: {800, 480},
	11: {1024, 768},
}

type Video struct {
	Stream      *mjpeg.Stream
	cam         *webcam.Webcam
	height      uint32
	width       uint32
	pixelFormat webcam.PixelFormat
	stop        chan struct{}
	fps         uint
	capStatus   bool
	vidDev      string
}

func NewVideo(pixelFormat webcam.PixelFormat, w uint32, h uint32, fps uint, vidDev string) *Video {
	return &Video{
		pixelFormat: pixelFormat,
		height:      h,
		width:       w,
		stop:        make(chan struct{}),
		fps:         fps,
		capStatus:   false,
		Stream:      mjpeg.NewStream(),
		vidDev:      vidDev,
	}
}

func (s *Video) SetResMode(i int) {
	s.SetRes(uint32(CamResolutions[i][0]), uint32(CamResolutions[i][1]))
}

func (s *Video) SetRes(w uint32, h uint32) {
	s.height = h
	s.width = w
}

func (s *Video) SetFPS(fps uint) {
	s.fps = fps
}

func (s *Video) StartVideoStream() error {
	cam, err := webcam.Open(s.vidDev)
	if err != nil {
		return err
	}

	format_desc := cam.GetSupportedFormats()

	f, w, h, err := cam.SetImageFormat(s.pixelFormat, s.width, s.height)
	if err != nil {
		return err
	}

	log.Printf("Resulting image format: %s %dx%d\n", format_desc[f], w, h)

	s.cam = cam

	if !s.capStatus {
		go s.startStreamer()
		return nil
	}
	log.Printf("Video capture already running")
	return nil
}

func (s *Video) StopVideoStream() error {
	if s.capStatus {
		s.stop <- struct{}{}
		if err := s.cam.StopStreaming(); err != nil {
			return err
		}
		if err := s.cam.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Video) startStreamer() {

	// Since the ReadFrame is buffered, trying to read at FPS results in delay.
	fpsTicker := time.NewTicker(time.Duration(1000/s.fps) * time.Millisecond)

	if err := s.cam.StartStreaming(); err != nil {
		log.Printf("Failed to start stream:%v", err)
		return
	}

	s.capStatus = true
	log.Print("Started Video Capture")

	frame := []byte{}
	buf := &bytes.Buffer{}
	for {
		select {
		case <-s.stop:
			log.Print("Stopped Video Capture")
			s.capStatus = false
			return

		default:
			if err := s.cam.WaitForFrame(5); err != nil {
				log.Printf("Failed to read webcam:%v", err)
			}
			var err error
			frame, err = s.cam.ReadFrame()
			if err != nil || len(frame) == 0 {
				log.Printf("Failed tp read webcam frame:%v or frame size 0", err)
			}

			// Convert frame to JPEG.
			buf, err = convertJPEG(frame, s.width, s.height)
			if err != nil {
				log.Printf("conver err %s", err)
			}

		case <-fpsTicker.C:
			s.Stream.UpdateJPEG(buf.Bytes())
		}
	}
}

func convertJPEG(frame []byte, w, h uint32) (*bytes.Buffer, error) {
	var img image.Image

	yuyv := image.NewYCbCr(image.Rect(0, 0, int(w), int(h)), image.YCbCrSubsampleRatio422)

	for i := range yuyv.Cb {
		ii := i * 4
		yuyv.Y[i*2] = frame[ii]
		yuyv.Y[i*2+1] = frame[ii+2]
		yuyv.Cb[i] = frame[ii+1]
		yuyv.Cr[i] = frame[ii+3]
	}
	img = yuyv

	//convert to jpeg
	buf := &bytes.Buffer{}
	if err := jpeg.Encode(buf, img, nil); err != nil {
		return nil, err
	}

	return buf, nil
}
