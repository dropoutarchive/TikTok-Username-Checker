package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/admin100/util/console"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpproxy"
)

var (
	avaliable int
	taken     int

	start int
	end   int

	usernames []string
	proxies   []string
	threads   int

	green = "\x1b[38;5;77m"
	red   = "\x1b[38;5;167m"
	reset = "\x1b[0m"
)

func main() {
	clear()
	rand.Seed(time.Now().UnixNano())
	console.SetConsoleTitle("TikTok Username Checker")
	wg := sync.WaitGroup{}

	f, err := os.Open("assets/usernames.txt")

	if err != nil {
		fmt.Printf("[%s%s%s] Failed to read assets/usernames.txt\n", red, time.Now().Format("15:04:05"), reset)
		return
	}

	s := bufio.NewScanner(f)
	for s.Scan() {
		usernames = append(usernames, s.Text())
	}

	f, err = os.Open("assets/proxies.txt")

	if err != nil {
		fmt.Printf("[%s%s%s] Failed to read assets/proxies.txt\n", red, time.Now().Format("15:04:05"), reset)
		return
	}

	s = bufio.NewScanner(f)
	for s.Scan() {
		proxies = append(proxies, s.Text())
	}

	fmt.Printf("[\x1b[38;5;63m%s\x1b[0m] Threads\x1b[38;5;63m>\x1b[0m ", time.Now().Format("15:04:05"))
	fmt.Scanln(&threads)

	clear()
	goroutines := make(chan struct{}, threads)
	start = int(time.Now().Unix())
	go background()

	for i := 0; i < len(usernames); i++ {
		wg.Add(1)
		go func(username string) {
			defer wg.Done()
			goroutines <- struct{}{}
			check(username)
			<-goroutines
		}(usernames[i])
	}
	wg.Wait()
	end = int(time.Now().Unix())

	fmt.Println()
	fmt.Println("Avaliable\x1b[38;5;63m:\x1b[0m", avaliable)
	fmt.Println("Taken\x1b[38;5;63m:\x1b[0m", taken)
	fmt.Printf("Took\x1b[38;5;63m:\x1b[0m %ds\n", (end - start))
	time.Sleep(3 * time.Millisecond)

}

func background() {
	for {
		console.SetConsoleTitle(fmt.Sprintf("TikTok Username Checker - Checked %d/%d", avaliable, (avaliable + taken)))
		time.Sleep(500)
	}
}

func check(username string) {
	if strings.Contains(username, "-") || strings.Contains(username, " ") {
		taken++
		fmt.Printf("[%s%s%s] Invalid (%s%s%s)\n", red, time.Now().Format("15:04:05"), reset, red, username, reset)
		return
	}

	req := fasthttp.AcquireRequest()
	res := fasthttp.AcquireResponse()

	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(res)

	req.Header.SetMethod("GET")
	req.SetRequestURI("https://t.tiktok.com/node/share/user/@" + username)
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.127 Safari/537.36")

	proxy := proxies[rand.Intn(len(proxies))]
	client := &fasthttp.Client{
		Dial:           fasthttpproxy.FasthttpHTTPDialer(proxy),
		ReadBufferSize: 50_000,
	}

	err := client.Do(req, res)
	if err != nil {
		return
	}

	body := string(res.Body())

	if strings.Contains(body, "10202") {
		avaliable++
		fmt.Printf("[%s%s%s] Avaliable (%s%s%s)\n", green, time.Now().Format("15:04:05"), reset, green, username, reset)
		file, _ := os.OpenFile("avaliable.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		file.Write([]byte(username + "\n"))
	} else if strings.Contains(body, "10221") || strings.Contains(body, "10222") {
		taken++
		fmt.Printf("[%s%s%s] Banned (%s%s%s)\n", red, time.Now().Format("15:04:05"), reset, red, username, reset)
	} else {
		taken++
		fmt.Printf("[%s%s%s] Taken (%s%s%s)\n", red, time.Now().Format("15:04:05"), reset, red, username, reset)
	}
}

func clear() {
	c := exec.Command("cmd", "/c", "cls")
	c.Stdout = os.Stdout
	c.Run()
}
