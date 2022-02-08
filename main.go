package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	var (
		vidDevice = flag.String("video_device", "/dev/video0", "Video device for teslong")
		vidH      = flag.Int("video_height", 240, "video height")
		vidW      = flag.Int("video_width", 320, "video width")
		vidFrame  = flag.Int("video_frame_rate", 2, "video frame rate")
		hostPort  = flag.String("host_port", ":8888", "HostPort")
	)
	flag.Parse()

	vid := NewVideo(YUYV422, uint32(*vidW), uint32(*vidH), uint(*vidFrame), *vidDevice)
	http.Handle("/", vid.Stream)

	if err := vid.StartVideoStream(); err != nil {
		log.Fatalf("Failed start %v", err)
	}
	if err := http.ListenAndServe(*hostPort, nil); err != nil {
		log.Fatalf("Fatal error starting http:%v", err)
	}

}
