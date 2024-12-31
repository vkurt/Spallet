package main

// import (
// 	"bytes"
// 	"image"
// 	"log"

// 	"fyne.io/fyne/v2/canvas"
// 	"fyne.io/fyne/v2/container"
// 	"fyne.io/fyne/v2/widget"
// 	"github.com/tuotoo/qrcode"
// 	"gocv.io/x/gocv"
// )

// func scanQr() { //Sadly this is not possible with current gocv package because it didnot supports any mobile, only option using gomobile package and using all go code like backend code
// 	// Initialize a
// 	w := spallet.NewWindow("QR Code Scanner")

// 	// Open the default camera
// 	webcam, err := gocv.VideoCaptureDevice(0)
// 	if err != nil {
// 		log.Fatalf("Error opening video capture device: %v", err)
// 	}
// 	defer webcam.Close()

// 	// Create an image canvas to display the captured frame
// 	img := gocv.NewMat()
// 	defer img.Close()
// 	imgCanvas := canvas.NewImageFromImage(image.NewRGBA(image.Rect(0, 0, 640, 480)))
// 	imgCanvas.FillMode = canvas.ImageFillOriginal

// 	// Add image canvas to the window
// 	w.SetContent(container.NewVBox(imgCanvas))

// 	go func() {
// 		for {
// 			// Read frame from the camera
// 			if ok := webcam.Read(&img); !ok {
// 				log.Printf("Device closed")
// 				return
// 			}
// 			if img.Empty() {
// 				continue
// 			}

// 			// Convert frame to image.Image
// 			buf, err := gocv.IMEncode(".jpg", img)
// 			if err != nil {
// 				log.Printf("Error encoding frame: %v", err)
// 				continue
// 			}
// 			imageData, _, err := image.Decode(bytes.NewReader(buf.GetBytes()))
// 			if err != nil {
// 				log.Printf("Error decoding frame: %v", err)
// 				continue
// 			}

// 			// Update the image canvas with the captured frame
// 			imgCanvas.Image = imageData
// 			imgCanvas.Refresh()

// 			// Decode the QR code from the buffer
// 			qr, err := qrcode.Decode(bytes.NewReader(buf.GetBytes()))
// 			if err != nil {
// 				log.Printf("No QR code found: %v", err)
// 				continue
// 			}

// 			// Display the QR code content
// 			log.Printf("QR Code Content: %s", qr.Content)
// 			w.SetContent(container.NewVBox(
// 				imgCanvas,
// 				widget.NewLabel("QR Code Content: "+qr.Content),
// 			))
// 		}
// 	}()

// 	// Show and run the app
// 	w.Show()
// }
