package main

import (
	"fmt"
	"log"
	"time"
)

type Pump struct {
	Name  string
	Fills int
}

type Car struct {
	Name  string
	Fills int
}

func PumpProducer(npumps int, newpumps <-chan Pump) <-chan Pump {
	rc := make(chan Pump)
	go func() {
		for i := 0; i < npumps; i++ {
			pump := Pump{Name: fmt.Sprintf("Pump #%d", i)}
			rc <- pump
		}
		for {
			select {
			case nextpump := <-newpumps:
				log.Printf("%s is available!", nextpump.Name)
				rc <- nextpump
			}
		}
	}()

	return rc
}

func CarProducer(ncars int, newcars <-chan Car) <-chan Car {
	rc := make(chan Car)
	go func() {
		for i := 0; i < ncars; i++ {
			car := Car{Name: fmt.Sprintf("Car #%d", i)}
			rc <- car
		}
		for {
			select {
			case nextcar := <-newcars:
				log.Printf("%s has driven to back of the queue.\n", nextcar.Name)
				rc <- nextcar
			}
		}
	}()

	return rc
}

func main() {
	var newpump = make(chan Pump, 10)
	var newcar = make(chan Car, 4)

	pumpchan := PumpProducer(4, newpump)
	carchan := CarProducer(10, newcar)

	start := time.Now()
	for {

		pump := <-pumpchan
		car := <-carchan

		go func() {
			log.Printf("Pumping gas into %s (%d) at pump %s (%d).\n",
				car.Name, car.Fills, pump.Name, pump.Fills)

			time.Sleep(50 * time.Millisecond)

			/* Increment our fill counters. */
			pump.Fills++
			car.Fills++

			/* put pumps and cars back into the queue. */
			newpump <- pump
			newcar <- car
		}()

		if int64(time.Since(start)/time.Second) > 10 {
			log.Printf("Running for %s\n", time.Since(start))
			break
		}
	}

	fmt.Printf("Generating report:")

PL:
	for {
		select {
		case pump := <-pumpchan:
			fmt.Printf("%s - %d\n", pump.Name, pump.Fills)
		case pump := <-newpump:
			fmt.Printf("%s - %d\n", pump.Name, pump.Fills)
		default:
			fmt.Printf("Waiting...")
			break PL
		}
	}

CL:
	for {
		select {
		case car := <-carchan:
			fmt.Printf("%s - %d\n", car.Name, car.Fills)
		case car := <-newcar:
			fmt.Printf("%s - %d\n", car.Name, car.Fills)
		default:
			break CL
		}
	}

}
