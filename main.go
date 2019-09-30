package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/stianeikeland/go-rpio"
)

func main() {
	fmt.Printf("Starting traffic lights at %s\n", time.Now())

	if err := rpio.Open(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Get the pin for each of the lights
	redPin := rpio.Pin(2)
	yellowPin := rpio.Pin(3)
	greenPin := rpio.Pin(4)

	// Set the pins to output mode
	redPin.Output()
	yellowPin.Output()
	greenPin.Output()

	// Clean up on ctrl-c and turn lights out
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c

		fmt.Printf("Switching off traffic lights at %s\n", time.Now())

		redPin.Low()
		yellowPin.Low()
		greenPin.Low()

		os.Exit(0)
	}()

	defer rpio.Close()

	// Turn lights off to start.
	redPin.Low()
	yellowPin.Low()
	greenPin.Low()

	fmt.Printf("All traffic lights switched off at %s\n\n", time.Now())

	// A while true loop.
	for {
		// Red
		redPin.High()
		time.Sleep(time.Second * 2)

		// Yellow
		redPin.Low()
		yellowPin.High()
		time.Sleep(time.Second)

		// Green
		yellowPin.Low()
		greenPin.High()
		time.Sleep(time.Second * 2)

		// Yellow
		greenPin.Low()
		yellowPin.High()
		time.Sleep(time.Second * 2)

		// Yellow off
		yellowPin.Low()
	}

}
