// Copyright 2012 Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/google/gousb"
	"github.com/puhitaku/mtplvcap/logging"

	"golang.org/x/sync/errgroup"

	"github.com/puhitaku/mtplvcap/mtp"
	"github.com/puhitaku/mtplvcap/public"
)

func main() {
	host := flag.String("host", "localhost", "hostname: default = localhost, specify 0.0.0.0 for public access")
	port := flag.Int("port", 42839, "port: default = 42839")
	backendGo := flag.Bool("backend-go", false, "use gousb as a libusb wrapper (not recommended)")
	debug := flag.String("debug", "", "comma-separated list of debugging options: usb, data, mtp, server")
	serverOnly := flag.Bool("server-only", false, "serve frontend without opening a DSLR (for devevelopment)")
	vendorID := flag.String("vendor-id", "0x0", "VID of the camera to search (in hex), default=0x0 (all)")
	productID := flag.String("product-id", "0x0", "PID of the camera to search (in hex), default=0x0 (all)")
	maxResolution := flag.Bool("max-resolution", false, "change the resolution to the max (experimental)")

	flag.Parse()

	debugs := map[string]bool{}
	for _, s := range strings.Split(*debug, ",") {
		debugs[s] = true
	}

	logging.SetLogLevel(debugs["main"], debugs["usb"], debugs["mtp"], debugs["data"], debugs["server"])
	log := logging.GetLogger().Main

	vid, err := strconv.ParseInt(strings.ReplaceAll(*vendorID, "0x", ""), 16, 64)
	if err != nil {
		log.Fatalf("failed to parse VID: %s", err)
	}

	pid, err := strconv.ParseInt(strings.ReplaceAll(*productID, "0x", ""), 16, 64)
	if err != nil {
		log.Fatalf("failed to parse PID: %s", err)
	}

	var dev mtp.Device

	if *serverOnly {
		log.Info("server-only mode is activated, skipping USB initialization")
	} else {
		if *backendGo {
			ctx := gousb.NewContext()
			defer ctx.Close()

			devGo, err := mtp.SelectDeviceGoUSB(ctx, uint16(vid), uint16(pid))
			if err != nil {
				log.Fatalf("failed to detect MTP device: %s", err)
			}
			defer devGo.Close()
			dev = devGo
		} else {
			devDirect, err := mtp.SelectDeviceDirect(uint16(vid), uint16(pid))
			if err != nil {
				log.Fatalf("failed to detect MTP devices: %v", err)
			}
			defer devDirect.Close()
			dev = devDirect
		}

		if err = dev.Configure(); err != nil {
			log.Fatalf("configure failed: %v", err)
		}
	}

	eg, ctx := errgroup.WithContext(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	eg.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return nil
			case s := <-sigChan:
				log.Infof("caught signal: %s", s)
				return errors.New(s.String())
			}
		}
	})

	lvs := mtp.NewLVServer(ctx, dev, *maxResolution)
	eg.Go(lvs.Run)

	router := http.NewServeMux()
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		f, _ := public.Root.Open("/controller.html")
		_, _ = io.Copy(w, f)
	})
	router.HandleFunc("/view", func(w http.ResponseWriter, r *http.Request) {
		f, _ := public.Root.Open("/index.html")
		_, _ = io.Copy(w, f)
	})
	router.HandleFunc("/mjpeg", lvs.HandleMotionJPEG)
	router.HandleFunc("/snapshot", lvs.HandleSnapshot)
	router.HandleFunc("/stream", lvs.HandleStream)
	router.HandleFunc("/control", lvs.HandleControl)
	router.Handle("/assets/", http.FileServer(public.Root))

	srv := http.Server{
		Addr:    fmt.Sprintf("%s:%d", *host, *port),
		Handler: logging.HTTPLogHandler(router),
	}

	eg.Go(func() error {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Errorf("http server returned an error: %s", err)
			return err
		}
		return nil
	})

	eg.Go(func() error {
		select {
		case <-ctx.Done():
		}
		ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
		return srv.Shutdown(ctx)
	})

	log.Info("started")

	err = eg.Wait()
	if err != nil {
		log.Error(err)
	}
}
