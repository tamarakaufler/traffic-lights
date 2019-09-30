package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	rpio "github.com/stianeikeland/go-rpio/v4"
)

func main() {
	fmt.Printf("Starting traffic lights at %s\n", time.Now())

	if err := rpio.Open(); err != nil {

		fmt.Printf("Cannot access GPIO: %s\n", time.Now())

		fmt.Println(err)
		os.Exit(1)
	}

	// Get the pin for each of the lights
	redPin := rpio.Pin(2)
	yellowPin := rpio.Pin(3)
	greenPin := rpio.Pin(4)

	fmt.Printf("GPIO input set up: %s\n", time.Now())

	// Set the pins to output mode
	redPin.Output()
	yellowPin.Output()
	greenPin.Output()

	fmt.Printf("GPIO output set up: %s\n", time.Now())

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
		fmt.Println("\tSwitching lights on and off")

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
