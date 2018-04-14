package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

func bytesToPowers(b int64) string {
	names :=
		[]string{"b", "kb", "mb", "gb", "tb", "pb", "eb", "zb", "yb"}
	l := len(names)
	i := 0

	res := float64(b)
	for res >= 1024 {
		if i == l-1 {
			break
		}

		res /= 1024
		i++
	}
	size := ""
	if i == 0 {
		size = strconv.FormatInt(int64(res), 10)
	} else {
		size = strconv.FormatFloat(res, 'f', 2, 64)
	}
	return size + names[i]
}

func worker(jobs <-chan string, done *sync.WaitGroup) {
	defer done.Done()

	for url := range jobs {
		path := strings.Split(url, "/")
		lastPart := path[len(path)-1]
		p, err := os.Getwd()
		if err != nil {
			fmt.Println("\x1b[31m[Error]\x1b[0m with path", lastPart)
			continue
		}
		out, err := os.Create(p + "/" + lastPart)
		if err != nil {
			fmt.Println("\x1b[31m[Error]\x1b[0m creating file", lastPart)
			continue
		}
		defer out.Close()

		resp, err := http.Get(url)
		if err != nil {
			fmt.Println("\x1b[31m[Error]\x1b[0m getting", url)
			continue
		}
		defer resp.Body.Close()

		n, err := io.Copy(out, resp.Body)
		if err != nil {
			fmt.Println("\x1b[31m[Error]\x1b[0m writing to", lastPart)
			continue
		}
		fmt.Println("\x1b[32m[Ok]\x1b[0m", lastPart, bytesToPowers(n))
	}
}

func main() {
	argsWithoutProg := os.Args[1:]
	if len(argsWithoutProg) < 1 {
		return
	}
	filename := argsWithoutProg[0]

	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}

	lines := strings.Split(string(content), "\n")

	jobs := make(chan string, 100)
	var done sync.WaitGroup

	for w := 0; w < 10; w++ {
		done.Add(1)
		go worker(jobs, &done)
	}

	for _, j := range lines {
		if j == "" {
			continue
		}
		jobs <- j
	}
	close(jobs)

	done.Wait()
}
