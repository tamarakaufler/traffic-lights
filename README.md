# traffic-lights

Activating containerized traffic lights (in Go) in Raspberry Pi Kubernetes cluster

Running containerized traffic lights in Raspberry Pi Kubernetes cluster
My cluster setup

The incentive behind this exercise was to understand how to set up deployment of an application talking to Raspberry Pi GPIOs when running the application in a container within a Kubernetes cluster.

Hardware

four Raspberry Pi 4 Model B (4Gb of memory)
4 microSD cards (3 x 32Gb, 1 x 64Gb):
Samsung EVO Plus 32 GB microSDHC 
SanDisk Ultra 32GB microSDHC Memory Card 
5 port USB power supply (PoE hat currently too expensive at £18 per one board)
5 port Ethernet switch (no PoE functionality):
TTP-Link LS1005G 5-Port Desktop/Wallmount Gigabit Ethernet Switch
4 USB cables (the best speed you can afford)
4 Ethernet cables
Multi Cable SLIM FLAT 1m Cat6 RJ45 Ethernet Network Patch Lan cable
2 traffic lights
3 lights:
3-Piece Set Ky 009 5050 3 Colour SDM RGB 3 Color LED Module for Arduino
3 GPIO header extensions:
T-Type GPIO Extension Board Module with 40 Pin Rainbow Ribbon Cable
Cluster tower:
MakerFun Pi Rack Case for Raspberry Pi 4 Model B,Raspberry Pi 3 B+ Case with Cooling Fan and Heatsink, 4 Layers Acrylic Case Stackable
Software
Balena Etcher
Raspbian lite (at the time of writing buster)
k3s or full blown k8s
k3s lightweight kubernetes setup:
either manually
using k3sup
k8s set up using kubeadmin
You will not need it for this project, but if you want to set up a Raspberry Pi Go development environment, it is possible to download a Go binary for the ARM architecture. uname -a command tells what architecture  the board has:

armv6l, armv7l .......... go1.13.1.linux-armv6l.tar.gz (at the time of writing), for Raspbian OS

arm64 ................... go1.13.1.linux-arm64.tar.gz, for other 64bit OSes

There are articles documenting setup of a Raspberry Pi Kubernetes cluster (see Links), so I shall not detail how I went about it here. Perhaps YARPCA in the future. One thing I found I needed to do to be able to access the cluster locally using the kubectl command, was to set the KUBECONFIG environment variable, even if I had the right (Raspberry Pi cluster) context in the default file ~/.kube/config. This was not obvious given the Kubernetes documentation.

I tried three ways of Kubernetes installation: using kubeadmin, k3sup and manual installation of k3s. I have now two clusters, one with the full blown Kubernetes container management system, one with the lighter k3s.

Each of the Raspberry Pi boards has a traffic light or an LED light (3 changing colours, ie connects to 3 pins + ground) connected to its GPIOs.
Fun with traffic lights

The traffic-lights Go code, Dockerfile and kubernetes manifests can be found on github.

main.go
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

 // Opens memory range for GPIO access in /dev/mem
 if err := rpio.Open(); err != nil {

  fmt.Printf("Cannot access GPIO: %s\n", time.Now())

  fmt.Println(err)
  os.Exit(1)
 }

 // Get the pin for each of the lights (refers to the bcm2835 layout)
 redPin := rpio.Pin(2)
 yellowPin := rpio.Pin(3)
 greenPin := rpio.Pin(4)

 fmt.Printf("GPIO pins set up: %s\n", time.Now())

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

 // Let's loop now ...
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


I went through several stages to learn and fully understand each scenario and to ensure all was working as supposed to.


Stage 1 - running the application directly on Raspberry Pi

After setting up the application dependencies using Go modules, I compiled a binary for the ARM architecture:

    GOOS=linux GOARCH=arm GOARM=7 go build -o trafficlights_arm7 .

Then transferred it to each of the boards to test the pins. The IP addresses are set up to be static in my router, eg the master is 192.168.1.92 etc. The hostnames are fixed as well, my master is raspberrypi-k3s-a.

    scp trafficlights_arm7 pi@192.168.1.92:.

I then sshed to each Raspberry Pi node and ran the traffic lights application

   ./trafficlights_arm7

All confirmed as working satisfactorily, I embarked on stage two, containerizing the traffic lights and running them on the Raspberry Pi nodes in containers.


Stage 2 - running the application in a container on Raspberry Pi

Dockerfile
FROM golang:1.13.1-buster as builder
WORKDIR /app
COPY . .

ENV GOARCH arm
ENV GOARM 7
ENV GOOS linux
RUN ["go", "build", "-o", "trafficlights", "."]

FROM scratch
WORKDIR /app
COPY --from=builder /app/trafficlights /app
CMD ["/app/trafficlights"]


I am using a two stage image build to have a lightweight final Docker image.
I create the image running the following command in the root of the traffic-lights git repo:

     docker build -t "forbiddenforrest/traffic-lights:0.1.0-armv7" .

I then log into my Docker registry (docker login) and push the image

     docker push forbiddenforrest/traffic-lights:0.1.0-armv7

Now I can ssh into one of my Raspberry Pi nodes and run a container based on the pushed traffic-lights image. The docker command is available when using the full version of kubernetes. When using k3s, docker is not available out of thew box as k3s uses containerd. docker can be downloaded separately (sudo apt-get install docker.io).
Manipulating the traffic lights using a containerized application did not prove straightforward. The container needs access to the Raspberry Pi node hardware, which is not given by default. I searched for a solution, came across some, but none worked for me. So I first needed to understand better how it works on the Raspberry Pi side and on the Docker side, then a spot of trial and error approach.

The result that worked for me:

     docker run --rm -it --device /dev/mem --device /dev/gpiomem  forbiddenforrest/traffic-lights:0.1.0-armv7

alternatively

     docker run --rm -it --device=/dev/mem --device=/dev/gpiomem  forbiddenforrest/traffic-lights:0.1.0-armv7

alternatively

     docker run --rm -it --device=/dev/mem:/dev/mem \
--device=/dev/gpiomem:/dev/gpiomem forbiddenforrest/traffic-lights:0.1.0-armv7


This (based on what I found when researching) did not work:

     docker run --rm -it --privileged forbiddenforrest/traffic-lights:0.1.0-armv7

     docker container run --rm -it --privileged --device=/dev/mem:/dev/mem --device=/dev/gpiomem:/dev/gpiomem -v /sys:/sys  forbiddenforrest/traffic-lights:0.1.0-armv7

     docker container run --rm -it --privileged  -v /sys:/sys forbiddenforrest/traffic-lights:0.1.0-armv7


All confirmed working, I moved to the next stage of running traffic-lights in a pod.

Stage 3 - running the application in a Pod in Raspberry Pi Kubernetes cluster

traffic_lights_pod.yaml
apiVersion: v1
kind: Pod
metadata:
  name: traffic-lights
  labels:
    app: traffic-lights
spec:
  securityContext:
    runAsUser: 1000
    runAsGroup: 997
    fsGroup: 15
  containers:
  - name: traffic-lights
    image: forbiddenforrest/traffic-lights:0.1.0-armv7
    securityContext:
      privileged: true

The securityContext needs to be set correctly both for the pod and the containers running in that pod:

runAsUser ..... 1000 (pi user)
runAsGroup ... 997  (group of /dev/mem, special character file which mirrors the main memory)
fsGroup .......... 15    (group of /dev/gpiomem, special character file which mirrors the memory associated with the GPIO device. Some volumes (storage) are owned and are writable by this GID.)

from
pi@raspberrypi-k3s-a:~ $ ls -l /dev/mem
crw-r----- 1 root kmem 1, 1 Oct  6 23:17 /dev/mem
pi@raspberrypi-k3s-a:~ $ ls -l /dev/gpiomem 
crw-rw---- 1 root gpio 247, 0 Oct  6 23:17 /dev/gpiomem
pi@raspberrypi-k3s-a:~ $ cat /etc/group |grep mem
kmem:x:15:
pi@raspberrypi-k3s-a:~ $ cat /etc/group |grep gpio
gpio:x:997:pi

Dealing with pods,  we are now fully dealing with the cluster, ie the traffic-lights application, when the pod is created, will be scheduled on one of the worker nodes. If the pod is killed and a new one created, it will be scheduled randomly by kubernetes. The master node is tainted not to be considered for scheduling.

To have a visual proof of the scheduling, I created a bash script for creating and deleting pods. Each of the worker nodes is connected to a light. When a pod gets deployed on a node, the lights connected to that particular Raspberry Pi board start working.

kubectl apply -f traffic_lights_pod.yaml
sleep 6
kubectl delete pod traffic-lights
kubectl apply -f traffic_lights_pod.yaml
sleep 6
kubectl delete pod traffic-lights
kubectl apply -f traffic_lights_pod.yaml
sleep 6
kubectl delete pod traffic-lights
kubectl apply -f traffic_lights_pod.yaml
sleep 6
kubectl delete pod traffic-lights

It took about 10 seconds to remove a deleted pod. If a pod is created directly in the cluster, kubernetes will not recreate it automatically when it is deleted. Let's deploy the traffic-lights using Kubernetes Deployment - my last Stage 4.


Stage 4 - running the application as a Deployment in Raspberry Pi Kubernetes cluster

traffic_lights_deploy.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: traffic-lights
  labels:
    app: traffic-lights
spec:
  replicas: 1
  selector:
    matchLabels:
      app: traffic-lights
  template:
    metadata:
      name: traffic-lights
      labels:
        app: traffic-lights
    spec:
      securityContext:
        runAsUser: 1000
        runAsGroup: 997
        fsGroup: 15
      containers:
      - name: traffic-lights
        image: forbiddenforrest/traffic-lights:0.1.0-armv7
        securityContext:
          privileged: true

The Deployment template contains the same Pod specification without apiVersion and kind information. The Deployment prescribes there should be one pod running at any time.

To see which worker the traffic-lights application is running on, I have the following script:

start_stop_deploy.sh
#!/bin/bash
echo "kubectl apply -f traffic_lights_deploy.yaml"
echo "then 10 rounds of pod deletion"
echo ""
echo "Start ..."
sleep 5
kubectl apply -f traffic_lights_deploy.yaml

for i in 1 2 3 4 5 6 7 8 9 10
do
    echo "round $i"
    echo "creating pod at `date`"
    sleep 4
    echo "deleting pod at `date`"
    kubectl delete pod -l 'app=traffic-lights'
done
kubectl delete deployments.apps/traffic-lights

echo "End ..."

It takes about 5 seconds for a new pod to start when one is deleted. A new pod is already in place while kubernetes is tidying up the deleted one.
