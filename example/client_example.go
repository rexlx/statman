package example

import (
	"flag"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/rexlx/statman/client"
)

var (
	minNum   = flag.Int("min", 1e6, "Minimum number")
	maxNum   = flag.Int("max", 1e9, "Maximum number")
	host     = flag.String("url", "cobra.nullferatu.com", "URL to send the request to")
	port     = flag.Int("port", 20080, "Port to send the request to")
	filename = flag.String("filename", "mojo", "Filename to save the stats to")
	count    = flag.Int("count", 20, "amount of requests to send")
)

type App struct {
	Stop chan interface{}
	client.StatsEndPoint
	ID string `json:"id"`
}

func GenerateRandomEvenNumber(min, max int) int {
	// Seed the random number generator with the current time.
	// rand.Seed()
	x := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Generate a random number within the specified range.
	randomNumber := x.Intn(max-min+1) + min

	// If the number is odd, add 1 to make it even.
	if randomNumber%2 != 0 {
		randomNumber++
	}

	return randomNumber
}

func GenerateRandomOddNumber(min, max int) int {
	// Seed the random number generator with the current time.
	// rand.Seed()
	x := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Generate a random number within the specified range.
	randomNumber := x.Intn(max-min+1) + min

	// If the number is even, add 1 to make it odd.
	if randomNumber%2 == 0 {
		randomNumber++
	}

	return randomNumber
}

func IsPrime(n int) bool {
	if n <= 1 {
		return false
	}

	for i := 2; i <= n/2; i++ {
		if n%i == 0 {
			return false
		}
	}

	return true
}

func Work() (int, bool) {
	x := GenerateRandomEvenNumber(*minNum, *maxNum)
	y := GenerateRandomOddNumber(*minNum, *maxNum)
	z := x + y
	return z, IsPrime(z)
}

func (a *App) Jiggle(n int) {
	var wg sync.WaitGroup
	defer func(t time.Time) {
		fmt.Println("Jiggle took: ", time.Since(t))
	}(time.Now())
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(n int, wg *sync.WaitGroup) {
			defer wg.Done()
			z, isPrime := Work()
			if isPrime {
				a.PostStatHTTP(
					client.Stat{
						ID:    "true",
						Value: float64(z),
						Time:  time.Now(),
					})
			} else {
				a.PostStatHTTP(
					client.Stat{
						ID:    "false",
						Value: float64(z),
						Time:  time.Now(),
					})
			}
		}(i, &wg)
	}

	wg.Wait()

}

func main() {
	flag.Parse()
	app := &App{
		Stop: make(chan interface{}),
		StatsEndPoint: client.StatsEndPoint{
			Host:     *host,
			Port:     *port,
			FileName: *filename,
		},
	}
	app.Jiggle(*count)
}
