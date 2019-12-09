package main

import (
	"cloud.google.com/aoc2019/day9/intcode"
	"flag"
	"fmt"
	"io/ioutil"
	"strconv"
	"sync"
)

func main() {
	var (
		vm *intcode.VM
		wg sync.WaitGroup
	)
	runmode := flag.Int("mode", 1, "run mode")
	flag.Parse()
	data, err := ioutil.ReadFile("boostPgm.dat")
	if err != nil {
		panic(err)
	}
	pgm := intcode.Compile(string(data))
	input := make(chan int, 2)
	output := make(chan int)
	input <- *runmode
	vm = intcode.NewVM(1, pgm, input, output)
	outstr := make([]string, 0)
	done := make(chan struct{})
	go func() {
		for o := range output {
			outstr = append(outstr, strconv.Itoa(o))
		}
		close(done)
	}()
	wg.Add(1)
	go func() {
		vm.Pgm.Debug(false)
		if err = vm.ExecPgm(); err != nil {
			panic(err)
		}
		wg.Done()
	}()
	wg.Wait()
	close(output)
	<- done
	fmt.Println(outstr)
}
