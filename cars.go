package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

/*
 * A Worker is an interface that represents either a pump or a car.
 *
 * We'll push instances of concrete types the implement Worker into our car and
 * pump queue.
 *
 */
type Worker interface {
	DoWork() bool
}

type Car struct {
	Name     string        /* name of car. */
	FillRate time.Duration /* time it takes to fill tank. */
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
	 * pumps really don't do anything.  we just need to have an available pump at
	 * the same time we have a car.
	 */
	pump.Pumps++
	return true
}

func (pump Pump) Report() {
	fmt.Printf("%s pumped %d times.\n", pump.Name, pump.Pumps)
}

func (car *Car) DoWork() bool {
	time.Sleep(car.FillRate * time.Millisecond)
	car.Fills++
	return true
}

func (car Car) Report() {
	fmt.Printf("%s filled %d times with a %d fill rate.\n", car.Name, car.Fills, car.FillRate)
}

func (car Car) GetName() string {
	return car.Name

}

func main() {
	npumps := flag.Int("pumps", 4, "Number of pumps.")
	ncars := flag.Int("cars", 10, "Number of cars.")
	duration := flag.Int("duration", 30, "Duraton of simulation")
	fillrate := flag.Int("fillrate", 50, "Fill rate in milliseconds/tank.  0 for random.")
	flag.Parse()

	runtime := time.Duration(*duration - 1)
	/* In both loops below, we have to pass pointers to our concrete types (Pump
	 * and Car).
	 *
	 *  https://golang.org/ref/spec#Method_sets
	 */

	/* Make a pump channel and prime it with new pumps. */
	pumpq := make(chan Worker, *npumps)
	for i := 0; i < *npumps; i++ {
		pumpq <- &Pump{Name: fmt.Sprintf("pump-%03d", i)}
	}

	/* Make a car channel and prime it with new cars. */
	carq := make(chan Worker, *ncars)
	for i := 0; i < *ncars; i++ {
		var d int
		if *fillrate == 0 {
			d = rand.Intn(5000)
		} else {
			d = *fillrate
		}
		carq <- &Car{Name: fmt.Sprintf("vehicle-%03d", i), FillRate: time.Duration(d)}
	}

	start := time.Now()

	var wg sync.WaitGroup

	/*
	 * The algorithm is pretty simple:  We have two channels (pumpq and carq)
	 * that have a backlog large enough to hold all of our pumps and cars.  We
	 * use these channels to queue those items.
	 *
	 * The program blocks until there is both a car and a pump available on the
	 * queue.	When there is, they are dequed and their their work methods are run
	 * concurrently.
	 *
	 * In this way, we use channels to send data and syncronize the program.
	 */

	log.Printf("Statring...")
	for time.Since(start)/time.Second > runtime {
		/* block until both a pump and a car is ready. */
		pump := <-pumpq
		car := <-carq

		go func() {
			wg.Add(1)
			defer wg.Done()

			log.Printf("%s filling at %s.", car.(*Car).GetName(), pump.(*Pump).GetName())
			pump.DoWork()
			car.DoWork() /* this will time.Sleep(). */

			/* put pump and car back into the queue. */
			pumpq <- pump
			carq <- car
		}()
	}

	log.Printf("Simulated for %s\n", time.Since(start))

	/* Wait for all remaining goroutines to finish. */
	log.Printf("Waiting for remaining go routines to complete...")
	wg.Wait()
	log.Printf("Simulation complete!")

	/* Read all pumps from pumpq chanel and report usage. */
	close(pumpq)
	for p := range pumpq {
		p.(*Pump).Report()
	}

	/* Read all cars from carq chanel and report usage. */
	close(carq)
	for c := range carq {
		c.(*Car).Report()
	}
}
