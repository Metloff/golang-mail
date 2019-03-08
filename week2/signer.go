package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

func main() {
}

// ExecutePipeline - конвеер, который выполняет функции которые поданы на вход.
func ExecutePipeline(jobs ...job) {
	wg := &sync.WaitGroup{}
	in := make(chan interface{}, 101)
	out := make(chan interface{}, 101)

	// Вычисления по цепочке.
	for _, j := range jobs {
		wg.Add(1)
		go runJob(wg, j, in, out)
		in = out
		out = make(chan interface{}, 101)
	}

	wg.Wait()
	close(out)
}

func runJob(wg *sync.WaitGroup, j job, in, out chan interface{}) {
	defer wg.Done()
	j(in, out)
	close(out)
}

func SingleHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}

	for value := range in {
		data, ok := value.(string)
		if !ok {
			data = strconv.Itoa(value.(int))
		}

		wg.Add(1)
		md := DataSignerMd5(data)

		go singleHash(wg, data, md, out)
	}

	wg.Wait()
}

func singleHash(wg *sync.WaitGroup, data, md string, out chan interface{}) {
	defer wg.Done()

	hash1 := make(chan string)
	hash2 := make(chan string)

	go func() {
		hash2 <- DataSignerCrc32(md)
	}()

	go func() {
		hash1 <- DataSignerCrc32(data)
	}()

	out <- fmt.Sprintf("%v~%v", <-hash1, <-hash2)
}

func MultiHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	for v := range in {
		wg.Add(1)
		go multiHash(wg, v.(string), out)
	}

	wg.Wait()
}

func multiHash(wg *sync.WaitGroup, data string, out chan interface{}) {
	defer wg.Done()

	var result string
	wg2 := &sync.WaitGroup{}
	resArr := make([]string, 7)
	for i := 0; i < 6; i++ {
		wg2.Add(1)
		go func(i int) {
			resArr[i] = DataSignerCrc32(fmt.Sprintf("%v%v", i, data))
			wg2.Done()
		}(i)
	}

	wg2.Wait()
	for _, res := range resArr {
		result += res
	}

	out <- result
}

func CombineResults(in, out chan interface{}) {
	var result []string

	for val := range in {
		result = append(result, val.(string))
	}

	sort.Strings(result)
	out <- strings.Join(result, "_")
}
