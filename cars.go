package main

import (
	"fmt"
	"log"
	"time"
)

type Worker interface {
	DoWork() bool
	Init(string)
	GetName() string
	Report()
}

type Car struct {
	Name     string        /* name of car. */
	FillTime time.Duration /* amount of time it takes to fill tank. */
	Fills    int           /* number of times a car has been filled. */
}

type Pump struct {
	Name  string /* name of pump. */
	Pumps int    /* number of times it has filled a tank. */
}

func (pump Pump) GetName() string {
	return pump.Name
}

func (pump *Pump) DoWork() bool {
	/*
	 * pumpq really don't do anything.  we just need to have an available pump at
	 * the same time we have a car.
	 */
	pump.Pumps++
	log.Printf("START PUMPING: %s", pump.Name)
	return true
}

func (pump Pump) Report() {
	fmt.Printf("%s pumped %d times.\n", pump.Name, pump.Pumps)
}

func (pump *Pump) Init(name string) {
	pump.Name = name
}

func (car *Car) DoWork() bool {
	log.Printf("START FILLING: %s", car.Name)
	time.Sleep(car.FillTime * time.Millisecond)
	log.Printf("END FILLING: %s", car.Name)
	log.Printf("%s leaving pump.", car.Name)
	car.Fills++
	return true
}

func (car Car) Report() {
	fmt.Printf("%s filled %d times.\n", car.Name, car.Fills)
}

func (car Car) GetName() string {
	return car.Name

}

func (car *Car) Init(name string) {
	car.Name = name
}

func main() {
	npumps := 4
	ncars := 10
	runtime := time.Duration(29)
	/* In both loops below, we have to pass pointers to our concrete types (Pump
	 * and Car).
	 *
	 *  https://golang.org/ref/spec#Method_sets
	 */

	/* Make a pump channel and prime it with new pumps. */
	pumpq := make(chan Worker, npumps)
	for i := 0; i < npumps; i++ {
		pumpq <- &Pump{Name: fmt.Sprintf("Pump #%d", i)}
	}

	/* Make a car channel and prime it with new cars. */
	carq := make(chan Worker, ncars)
	for i := 0; i < ncars; i++ {
		carq <- &Car{Name: fmt.Sprintf("Vehicle #%d", i), FillTime: 50}
	}

	start := time.Now()
	for {
		/* block until both a pump and a car is ready. */
		pump := <-pumpq
		car := <-carq

		go func() {
			log.Printf("About to fill %s from %s", car.GetName(), pump.GetName())
			pump.DoWork()
			car.DoWork() /* this will time.Sleep(). */

			/* put pump and car back into the queue. */
			pumpq <- pump
			carq <- car
		}()

		if time.Since(start)/time.Second > runtime {
			log.Printf("Running for %s\n", time.Since(start))
			break
		}
	}

	/* Wait for all remaining goroutines to finish. */
	time.Sleep(500 * time.Millisecond)

	/* Read all pumps from pumpq chanel and report usage. */
pr:
	for {
		select {
		case pump := <-pumpq:
			pump.Report()
		default:
			break pr
		}
	}

	/* Read all cars from carq chanel and report usage. */
vr:
	for {
		select {
		case car := <-carq:
			car.Report()
		default:
			break vr
		}
	}
}
