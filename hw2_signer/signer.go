package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

func ExecutePipeline(jobs ...job) {

	wg := &sync.WaitGroup{}
	in := make(chan interface{})
	for ind, _func := range jobs{
		out := make(chan interface{})
		wg.Add(1)
		go func(in, out chan interface{}, wg *sync.WaitGroup, f job, index int) {
			fmt.Println("GO func started ", index)
			defer wg.Done()
			defer func(out chan interface{}) {
				close(out)
				fmt.Println("Out chan closed ", index)
			}(out)
			f(in, out)
		}(in, out, wg, _func, ind)

		in = out
	}
	wg.Wait()
}

func routineSingleHash(data interface{}, out chan interface{}, wg *sync.WaitGroup, index int, md5 string) {
	defer wg.Done()
	fmt.Println("SingleHash routine ", index)

	c1 := make(chan string)
	c2 := make(chan string)
	var pure, compose string

	go func() {
		c1 <- DataSignerCrc32(md5)
	}()

	go func() {
		c2 <- DataSignerCrc32(strconv.Itoa(data.(int)))
	}()

	compose = <-c1
	fmt.Println("SingleHash crc32(md5(data)) ", compose)
	pure = <-c2
	fmt.Println("SingleHash crc32(data) ", pure)

	out <- pure + "~" + compose
}

func SingleHash(in, out chan interface{}) {
	start := time.Now()
	wg := &sync.WaitGroup{}
	var i int

	for data := range in {
		wg.Add(1)
		_md5 := DataSignerMd5(strconv.Itoa(data.(int)))

		go routineSingleHash(data, out, wg, i, _md5)
		//routineSingleHash(data, out, wg, i)
		i++
	}
	wg.Wait()
	end := time.Since(start)
	fmt.Println("SingleHash time ", end)
}


func routineMultiHash(data interface{}, out chan interface{}, wg *sync.WaitGroup, index int) {
	start := time.Now()
	defer wg.Done()
	wgLocal := &sync.WaitGroup{}
	fmt.Println("MultiHash routine ", index, " started")
	buffer := make([]string, 6)

	for i := 0; i < 6; i++ {
		wgLocal.Add(1)

		go func(i int, sGroup *sync.WaitGroup, buffer []string) {
			defer sGroup.Done()
			//fmt.Println("crc32 MultiHash ", i)
			buffer[i] = DataSignerCrc32(strconv.Itoa(i) + data.(string))
		}(i, wgLocal, buffer)

	}
	wgLocal.Wait()
	out <- strings.Join(buffer, "")
	end := time.Since(start)
	fmt.Println("routineMultiHash time ", end)
}

func MultiHash(in, out chan interface{}) {
	start := time.Now()

	wg := &sync.WaitGroup{}
	var i int

	for data := range in {
		wg.Add(1)
		go routineMultiHash(data, out, wg, i)
		i++
	}
	wg.Wait()

	end := time.Since(start)
	fmt.Println("MultiHash time ", end)
}

func CombineResults(in, out chan interface{}) {
	buffer := []string{}
	var i int

	for data := range in {

		buffer = append(buffer, data.(string))
		i++
	}

	sort.Slice(buffer, func(i, j int) bool {
		return buffer[i] < buffer[j]
	})

	fmt.Println("Combine result func ended")
	fmt.Println(strings.Join(buffer[:], "_"))
	out <- strings.Join(buffer[:], "_")
}

