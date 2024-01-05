package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

/*
	👏1. SENDING OR RECEIVING CALL ON A CHANNEL ARE BLOCKING IN NATURE UNTIL SENDER FINDS A RECEIVER OR RECEIVER FINDS A SENDER
	👏2. MAIN SHOULD BLOCK TO ALLOW OTHER GO-ROUTINES TO EXECUTE (time.Sleep() || Stdin || Reading from Channel) WHICH ARE CREATED IN MAIN
			OTHERWISE THOSE GO-ROUTINES ARE ONLY CREATED BUT NEVER EXECUTED!
	👏3. DATA IN CHANNELS FLOW IN ORDER
	👏4. CONCURRENCY CAN LEAD TO PARALELLISM (DEPENDS  ON YOUR H/W)

	👍 DO NOT LOOP ON CHANNEL.
	👍 IF THERE IS NO DATA TO READ, It'LL WAIT FOR THE DATA THAT'S NEVER GOING TO ARRIVE.
	👍 IF I'M NOT CLOSING THE CHANNEL, SO I WANT TO MAKE SURE THAT I'm NOT READING DATA MORE TIMES THAN THERE COULD BE DATA IN THE CHANNEL.
	👍 WHY DON'T I CLOSE THE CHANNEL?
	👍 YOU CAN ONLY CLOSE THE CHANNEL ONCE! YOU CAN'T CLOSE THE CHANNEL THAT's ALREADY  BEEN CLOSED.
	👍 I HAVE 99 GO-ROUTINES? WHO CLOSES THE CHANNEL, WHICH ONE AMONG THEM ?
	👍 LAST PERSON ALWAYS TURNS THE LIGHTS OFF, BUT THERE IS NO WAY IN GO-ROUTINES TO KNOW EXACTLY WHICH ONE IS THE LAST?

	👍 SELECT:- CONTROL STRUCTURE THAT WORKS ON CHANNELS
	👍 USING SELECT I CAN LISTEN TO MORE THAN ONE CHANNEL SIMULTANEOUSLY
	👍 PROBLEM WITHOUT SELECT:- POOLING MEANS WAITING FOR THE INPUT FROM MULTIPLE CHANNELS SEQUENTIALLY
	👍 THE ISSUE HERE IS THAT IF ONE CHANNEL HAS DATA AVAILABLE MORE FREQUENTLY THAN THE OTHER, THE PROGRAM|LOOP
	MIGHT BE UNNECESSARILY DELAYED

	* 💎💎💎💎💎💎💎💎💎
	CONTEXT:- CONTEXT IS LIKE A BAG OR CONTAINER THAT HOLDS INFORMATION THAT IS SHARED BETWEEN DIFFERENT PARTS OF THE PROGRAM,
	ESPICIALLY WHEN IT COMES TO HANDLING A REQUEST. THIS INFORMATION CAN INCLUDE THINGS LIKE TIMEOUTS, CANCELLATION SIGNALS,
	AND OTHER DATA THAT IS SPECIFIC TO THAT REQUEST.
	EXAMPLE:- IMAGINE YOU ARE BUILDING A WEB SERVER THAT HANDLES A LOT OF INCOMING REQUESTS. EACH REQUEST HAS ITS OWN SPECIFIC
	NEEDS & REQUIREMENTS, SUCH AS DEADLINE FOR HOW LONG IT SHOULD TAKE TO COMPLETE. THE CONTEXT ALLOWS YOU TO KEEP TRACK OF THESE
	INDIVIDUAL REQUIREMENTS FOR EACH REQUEST AND MAKE SURE THAT THEY ARE HANDLED PROPERLY!

	CONTEXTs FORM IMMUTABLE TREE DATA STRUCTURE.

	🤷 context.WithCancel:- CREATES A CONTEXT THAT CAN BE CANCELLED MANUALLY BY CALLING THE CANCEL FUNCTION RETURNED BY WithCancel
				ctx, cancel := context.WithCancel(context.Background())
				cancel()

	🤷 context.WithTimeout:- IS A SPECIAL FORM OF WithCancel WHERE THE CONTEXT IS AUTOMATICALLY CANCELLED AFTER A SPECIFIED DURATION
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()

     	!GO-ROUTINES KEEP POINTERS TO THE SHARED CHANNELS, IF THE CHANNELS ARE NOT CLOSED IT LEADS TO GO-ROUTINE LEAKS.
	💎💎💎💎💎💎💎💎💎
	🤷
	?WAITGROUP:-
		💎 A WAITGROUP WAITS FOR A COLLECTION OF GOROUTINES TO FINISH
		💎 THE MAIN GOROUTINE CALLS ADD() TO SET THE NUMBER OF GOROUTINES TO WAIT FOR
		💎 THEN EACH OF THE GOROUTINES RUNS & CALLS DONE() WHEN FINISHED
		💎 AT THE SAME TIME, WAIT() CAN BE USED TO BLOCK UNITLL ALL GOROUTINES HAVE FINISHED

		 *we use a WaitGroup to coordinate the goroutines. The Add method increments the counter,
		 *and each goroutine calls Done when it finishes. The Wait method blocks until the counter becomes zero
		 *indicating that all goroutines have completed.
		 *WaitGroup avoids us to rely on a time.Sleep to give goroutines some time to finish, so we can conclude that
		 *WaitGroup is a synchronization mechanism

*/

/*
👏We can't take address of a MAP or SLICE entry, why?
Because slices can get reallocated & we are then storing some stale address, in case of maps, maps keep on changing
(rearranging themselves here and there i.e no order), so we too might keep pointing to some stale pointer variable.
AVOID:- &slice[0] or %map[key]
Always gonna see map[string]*Employee and not map[string]Employee

👏Do not capture refrence to a Loop variable

OOPS

👏IN GO WE CAN PUT METHODS ON ANY USER DECLARED TYPE
👏NOTE THAT I CAN ASSIGN ANYTHING TO THE INTERFACE THAT SATISFIES THE INTERFACE (ANY TYPE THAT HAS ALL INTERFACE METHODS)
A METHOD IS A FUNCTION ASSOCIATED WITH A TYPE
👏AN INTERFACE IS JUST LIKE A ABSTRACT PARENT CLASS (WHICH DEFINES THE BEHAVIOUR)
👏WE CAN ASSIGN TO INTERFACE ANY TYPE THAT IMPLEMENTS INTERFACE

IN COMPOSITION, FIELDS & METHODS ARE PROMOTED
WE CAN ALSO PROMOTE INTERFACE WITHIN A STRUCT
*/

// CONCURRENCY START

// 👏GOROUTINES:- 1. A goroutine is a lightweight thread managed by the Go runtime.
// 2. It's a function that runs concurrently with other goroutines in the same address space.

// 👏CHANNELS:- 1. Channels are communication pipes that allow one goroutine to send data to another goroutine.
// 2. Channels provide a safe way for goroutines to communicate and synchronize their execution.
// 3. You can send data into a channel from one goroutine and receive it in another.

// ?EXAMPLE 21 GENERATORS [PRIME NUMBERS FINDER] Fan In, Fan Out Pattern

// *generators are only going to produce the amount of data that we are going to take, not running infinitely!

func generator(done <-chan int, fn func() int) <-chan int {
	stream := make(chan int)

	go func() {
		defer close(stream)

		for {
			select {
			case <-done:
				return
			case stream <- fn():
				// fmt.Println("ticking")
			}
		}
	}()

	return stream
}

func primeFinder(done <-chan int, randIntStream <-chan int) <-chan int {
	isPrime := func(randomInt int) bool {
		for i := randomInt - 1; i > 1; i-- {
			if randomInt%i == 0 {
				return false
			}
		}
		return true
	}

	primes := make(chan int)

	go func() {
		defer close(primes)

		for {
			select {
			case <-done:
				return
			case randomInt := <-randIntStream:
				// fmt.Println(randomInt)
				if isPrime(randomInt) {
					primes <- randomInt
				}
			}
		}
	}()

	return primes
}

func take(done <-chan int, stream <-chan int, n int) <-chan int {
	taken := make(chan int)

	go func() {
		defer close(taken)

		for i := 1; i <= n; i++ {
			select {
			case <-done:
				return
			case taken <- <-stream:
			}
		}
	}()

	return taken
}

func fanIn[T any, K any](done <-chan K, channels ...<-chan T) <-chan T {
	var wg sync.WaitGroup

	fannedInStream := make(chan T)

	transfer := func(c <-chan T) {
		defer wg.Done()

		for p := range c {
			select {
			case <-done:
				return
			case fannedInStream <- p:
			}
		}
	}

	for _, c := range channels {
		wg.Add(1)
		go transfer(c)
	}

	go func() {
		wg.Wait()
		// When all the transfers are done close the channel
		close(fannedInStream)
	}()

	return fannedInStream
}

func main() {
	start := time.Now()

	done := make(chan int)
	defer close(done)

	randInt := func() int {
		return rand.Intn(100000000)
	}

	randomIntStream := generator(done, randInt)

	// !naive
	// primeStream := primeFinder(done, randomIntStream)
	// for p := range take(done, primeStream, 10) {
	// 	fmt.Println(p)
	// }

	// ?Fan Out Spinning primeFinder multiple times
	CPUCount := runtime.NumCPU()
	primeFinderChannels := make([]<-chan int, CPUCount)

	for i := 0; i < CPUCount; i++ {
		primeFinderChannels[i] = primeFinder(done, randomIntStream)
	}

	// ?Fan In
	fannedInStream := fanIn(done, primeFinderChannels...)
	for p := range take(done, fannedInStream, 5) {
		fmt.Println(p)
	}

	fmt.Println(time.Since(start))
}

// ?EXAMPLE 20 GENERATORS WITH INFINITE STREAM OF DATA

// func repeatFunc[T any, K any](done <-chan K, fn func() T) <-chan T {
// 	stream := make(chan T)

// 	go func() {
// 		for {
// 			select {
// 			case <-done:
// 				return
// 			case stream <- fn():
// 			}
// 		}
// 	}()
// 	return stream
// }

// func main() {
// 	done := make(chan int)

// 	randomNumber := func() int {
// 		return rand.Intn(100000)
// 	}

// 	for rand := range repeatFunc(done, randomNumber) {
// 		fmt.Println(rand)
// 	}
// }

// ?EXAMPLE 19 CONFINEMENT:- We confine each go-routine to a specific part of the shared resource

// func getEvenDate(n int) int {
// 	time.Sleep(time.Second * 2)
// 	return 2 * n
// }

// func processData(wg *sync.WaitGroup, result *int, n int) {
// 	defer wg.Done()

// 	even := getEvenDate(n)
// 	(*result) = even
// }

// func main() {
// 	start := time.Now()

// 	var wg sync.WaitGroup

// 	input := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
// 	result := make([]int, len(input))

// 	for i, n := range input {
// 		wg.Add(1)
// 		// ?confine each go-routine to a specifc memory addr not giving entire slice
// 		// ?each go-routine acts on a different memory address
// 		go processData(&wg, &result[i], n)
// 	}

// 	wg.Wait()

// 	fmt.Println("result:-", result)
// 	fmt.Println("time:-", time.Since(start))
// }

// ?EXAMPLE 18 LOCKING adds BOTTLENECK

// var lock sync.Mutex

// func getEvenDate(n int) int {
// 	time.Sleep(time.Second * 2)
// 	return 2 * n
// }

// func processData(wg *sync.WaitGroup, result *[]int, n int) {
// 	// !note that we have not used locking in optimal way, but this is just for example
// 	lock.Lock()
// 	defer wg.Done()

// 	even := getEvenDate(n)
// 	(*result) = append((*result), even) // *shared resource
// 	lock.Unlock()
// }

// func main() {
// 	start := time.Now()

// 	var wg sync.WaitGroup

// 	input := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
// 	result := []int{} // ?shared resource

// 	for _, n := range input {
// 		wg.Add(1)
// 		go processData(&wg, &result, n)
// 	}

// 	wg.Wait()

// 	fmt.Println("result:-", result)
// 	fmt.Println("time:-", time.Since(start))
// }

// !EXAMPLE 17 NON CONFINEMENT (WAITGROUPS)

// func processData(wg *sync.WaitGroup, result *[]int, n int) {
// 	defer wg.Done()

// 	even := n * 2
// 	(*result) = append((*result), even)
// }

// func main() {
// 	var wg sync.WaitGroup

// 	input := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
// 	result := []int{} // ?shared resource

// 	for _, n := range input {
// 		wg.Add(1)
// 		go processData(&wg, &result, n)
// 	}

// 	wg.Wait()

// 	fmt.Println("result:-", result)
// }

// !EXAMPLE 16 BEAUTIFUL RACE CONDITION EXAMPLE WITH 1000 GO-ROUTINES ON CRITICAL VARIABLE (READ MODIFY WRITE CYCLE)

// func decrementByOne(amount *int) {
// 	(*amount) -= 1
// }

// func main() {
// 	amount := 1000
// 	fmt.Println("amount at start", amount)

// 	for i := 1; i <= 1000; i++ {
// 		go decrementByOne(&amount)
// 		// decrementByOne(&amount)
// 	}

// 	fmt.Println("100 goroutines running")
// 	time.Sleep(5)
// 	fmt.Println("amount at end", amount)
// }

// ?EXAMPLE 15 PIPELINING (STAGES FOR DATA PROCESSING)

// type Array []int

// func (arr Array) sliceToChannel() <-chan int {
// 	out := make(chan int)
// 	// out := make(chan int, 10000000)

// 	go func() {
// 		for _, n := range arr {
// 			out <- n
// 		}
// 		close(out)
// 	}()
// 	return out
// }

// func square(in <-chan int) <-chan int {
// 	out := make(chan int)
// 	// out := make(chan int, 10000000)

// 	go func() {
// 		for n := range in {
// 			out <- n * n
// 		}
// 		close(out)
// 	}()
// 	return out
// }

// func main() {
// 	start := time.Now()
// 	fmt.Println("start")
// 	var nums Array
// 	for i := 1; i <= 10000000; i++ {
// 		nums = append(nums, i)
// 	}

// 	dataChannel := nums.sliceToChannel()
// 	sqChannel := square(dataChannel)

// 	for n := range sqChannel {
// 		fmt.Printf("%d  ", n)
// 	}

// 	end := time.Since(start)
// 	fmt.Println("end", end)
// }

// !EXAMPLE 14 (More about Select Usecase)

// func main() {
// 	fmt.Println("start")
// 	fastChannel := make(chan string)
// 	slowChannel := make(chan string)

// 	go func() {
// 		time.Sleep(time.Second * 2)
// 		slowChannel <- "Sending data to slow channel"
// 	}()

// 	go func() {
// 		time.Sleep(time.Second * 1)
// 		fastChannel <- "Sending data to fast channel"
// 	}()

// 	// fmt.Println("slow read first", <-slowChannel)
// 	// fmt.Println("fast read last", <-fastChannel)

// loop:
// 	for {
// 		select {
// 		case s := <-slowChannel:
// 			fmt.Println("reading from slow channel", s)
// 			break loop
// 		case f := <-fastChannel:
// 			fmt.Println("reading from fast channel", f)
// 		}
// 	}

// 	fmt.Println("end")
// }

// !EXAMPLE 13 SEQUENTIAL FILE PROCESSING:- Searching duplicate files based on their hash

// type pair struct {
// 	hash, path string
// }

// type filelist []string //slice of files

// type results map[string]filelist // hash -> [...files]

// func hashFile(path string) pair {
// 	file, err := os.Open(path)

// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer file.Close()

// 	hash := md5.New()

// 	if _, err := io.Copy(hash, file); err != nil {
// 		log.Fatal(err)
// 	}

// 	return pair{fmt.Sprintf("%x", hash.Sum(nil)), path}
// }

// // Take I/P directory and generates hash for every file inside the directory recursively
// func searchTree(dir string) (results, error) {
// 	hashes := make(results)

// 	// !The Walk function will recursively visit all files and directories under the specified root.
// 	// ! os.FileInfo object providing information about the file or directory

// 	err := filepath.Walk(dir, func(p string, fi os.FileInfo, err error) error {
// 		if err != nil {
// 			return err
// 		}
// 		if !fi.IsDir() {
// 			fileSize := int(fi.Size() / (1024 * 1024))
// 			if fileSize == 0 {
// 				fileSize = 1
// 			}
// 			fmt.Printf("size of %15v is %10d MB\n", fi.Name(), fileSize)
// 		}
// 		//? All empty Files will have same hashValue
// 		if fi.Mode().IsRegular() && fi.Size() > 0 {
// 			h := hashFile(p)
// 			hashes[h.hash] = append(hashes[h.hash], h.path)
// 		}
// 		return nil
// 	})

// 	return hashes, err
// }

// func main() {
// 	start := time.Now()

// 	if len(os.Args) < 2 {
// 		log.Fatal("missing parameters, provide directory name")
// 	}

// 	if hashes, err := searchTree(os.Args[1]); err == nil {
// 		// fmt.Println(hashes)
// 		for hash, files := range hashes {

// 			if len(files) > 1 {
// 				// !duplicate files
// 				fmt.Printf("#duplicate files (content) with hash %v are: [%v] \n", hash, strings.Join(files, ", "))
// 				fmt.Println()
// 			}
// 		}
// 	}
// 	fmt.Println(time.Since(start))
// }

// ?EXAMPLE 13 (We can loop over closed buffered channel)

// func main() {
// 	tick := time.NewTicker(time.Millisecond * 100).C
// 	stopper := time.After(time.Second * 2)

// 	fmt.Println("start")

// 	chars := []string{"a", "b", "c"}
// 	charChannel := make(chan string, 3)

// 	go func() {
// 		for _, char := range chars {
// 			time.Sleep(time.Millisecond * 300)
// 			charChannel <- char
// 		}
// 		close(charChannel)
// 		// charChannel <- "Haha" //! panic: send on closed channel
// 	}()

// loop:
// 	for {
// 		select {
// 		case <-stopper:
// 			fmt.Println("stop")
// 			break loop
// 		case <-tick:
// 			fmt.Println("tick")
// 		case c, ok := <-charChannel:
// 			if !ok {
// 				break loop
// 			}
// 			fmt.Println("value", c, ok)
// 		}
// 	}

// 	time.Sleep(time.Second * 2) // checking if there's any Go-routine leak
// 	fmt.Println("end")
// }

// ?EXAMPLE 12 RACE CONDITION

// var arr = make([]int, 5)

// func do(i int, ch chan<- *int) {
// 	ch <- &i
// 	fmt.Println("more++")
// 	i += 100 //! UNSAFE (Read Modify Write Cycle)
// }

// func main() {
// 	// ch := make(chan *int, 5)
// 	ch := make(chan *int)
// 	// stopper := time.After(time.Second * 1)
// 	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
// 	defer cancel()

// 	for i := 0; i < 5; i++ {
// 		go do(i, ch)
// 	}

// 	for {
// 		select {
// 		case <-ctx.Done():
// 			log.Fatal("exit")
// 		case v := <-ch:
// 			fmt.Println("value", *v)
// 		}
// 	}
// }

// ?EXAMPLE 11 CONTEXT WITH HTTP GET

// type result struct {
// 	url     string
// 	err     error
// 	latency time.Duration
// }

// func get(ctx context.Context, url string, ch chan<- result) {
// 	// INJECTING TIMEOUT IN THE HTTP GET REQUEST
// 	start := time.Now()
// 	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

// 	if resp, err := http.DefaultClient.Do(req); err != nil {
// 		ch <- result{url, err, 0}
// 	} else {
// 		t := time.Since(start).Round(time.Millisecond)
// 		ch <- result{url, nil, t}
// 		resp.Body.Close()
// 	}
// }

// func main() {
// 	results := make(chan result)
// 	list := []string{
// 		"https://amazon.com",
// 		"https://google.com",
// 		"https://nytimes.com",
// 		"https://youtube.com",
// 		"http://localhost:8000",
// 	}

// 	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
// 	defer cancel()

// 	for _, url := range list {
// 		go get(ctx, url, results)
// 	}

// 	for range list {
// 		r := <-results

// 		if r.err != nil {
// 			log.Printf("%-20s %s\n", r.url, r.err)
// 		} else {
// 			log.Printf("%-20s %s\n", r.url, r.latency)
// 		}
// 	}
// }

// ?EXAMPLE 10 CONTEXT [******************]

// func slowFetching(ctx context.Context) (int, error) {
// 	responseChannel := make(chan int)
// 	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*2000)
// 	defer cancel()

// 	go func() {
// 		time.Sleep(time.Millisecond * 2500)
// 		responseChannel <- 201
// 	}()

// 	select {
// 	case <-ctx.Done():
// 		return 0, fmt.Errorf("Slow api")
// 	case r := <-responseChannel:
// 		return r, nil
// 	}
// }

// func main() {
// 	ctx := context.Background()
// 	value, err := slowFetching(ctx)

// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	fmt.Println("value", value)
// }

// ?💎 EXAMPLE 09 CONTEXT
// 🤷 The context package in Golang provides a way to pass cancellation signals and deadlines to functions and goroutines

// func fetchThirdPartyStuffWhichCanBeSlow() (int, error) {
// 	time.Sleep(time.Millisecond * 3000)

// 	return 200, nil
// }

// type Response struct {
// 	value int
// 	err   error
// }

// func fetchUserDate(ctx context.Context, userID int) (int, error) {
// 	val := ctx.Value("foo")
// 	fmt.Println("value from context: ", val)
// 	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*2000) // timeout 2 seconds
// 	defer cancel()

// 	responseChannel := make(chan Response)

// 	go func() {
// 		val, err := fetchThirdPartyStuffWhichCanBeSlow()
// 		responseChannel <- Response{val, err}
// 	}()

// 	for {
// 		select {
// 		// Done is called when timeout is triggered
// 		case <-ctx.Done():
// 			return 0, fmt.Errorf("api took too long")
// 		case res := <-responseChannel:
// 			return res.value, res.err
// 		}
// 	}
// }

// func main() {
// 	start := time.Now()
// 	// ctx := context.Background()
// 	ctx := context.WithValue(context.Background(), "foo", "bar") // storing key value data in context
// 	userID := 10
// 	val, err := fetchUserDate(ctx, userID)

// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	fmt.Println("result: ", val)
// 	fmt.Println("took", time.Since(start))
// }

// ?EXAMPLE 08 SELECT || PERIODIC TASKS

// func main() {
// 	fmt.Println("start")
// 	tickRate := 2 * time.Second

// 	// sends the current time on the returned channel after duration is elasped
// 	doneChannel := time.After(5 * tickRate)
// 	// send the current time on the channel after each tick (PERIODIC TASK)
// 	periodicChannel := time.NewTicker(tickRate).C
// loop:
// 	for {
// 		select {
// 		case <-periodicChannel:
// 			// refresh cache after certain period
// 			log.Println("tick")
// 		case <-doneChannel:
// 			break loop
// 		}
// 	}
// 	fmt.Println("finish")
// }

// ?EXAMPLE 07 SELECT WITH WEB SERVICES
// CLOSE THE PROGRAM AFTER 5 SECONDS.
// INTENTIONALLY MY LOCAL SERVER IS TAKING 10 SECONDS
// WATCH FOR ANOTHER TIMER CHANNEL (simultaneously)
// 👀👀👀

// type result struct {
// 	url     string
// 	err     error
// 	latency time.Duration
// }

// func get(url string, ch chan<- result) {
// 	start := time.Now()

// 	if resp, err := http.Get(url); err != nil {
// 		ch <- result{url, err, 0} // 🌻writing to channel, blocking call until it finds some receiver
// 	} else {
// 		t := time.Since(start).Round(time.Millisecond)
// 		ch <- result{url, nil, t} // 🌻writing to channel, blocking call until it finds some receiver
// 		resp.Body.Close()
// 	}
// }

// func main() {
// 	// waits for duration to elapse & then sends the current time on returned channel!
// 	stopper := time.After(5 * time.Second)
// 	results := make(chan result)
// 	list := []string{
// "https://amazon.com",
// "https://google.com",
// "https://nytimes.com",
// "https://youtube.com",
// "http://localhost:8000",
// 	}

// 	for _, url := range list {
// 		go get(url, results)
// 	}

// 	for range list {
// 		// 🌻Reading from channel (blocking call, until some go-routine writes to the channel) , data flows in order

// 		// r := <-results
// 		select {
// 		case r := <-results:
// 			if r.err != nil {
// 				log.Printf("%-20s %s\n", r.url, r.err)
// 			} else {
// 				log.Printf("%-20s %s\n", r.url, r.latency)
// 			}
// 		case t := <-stopper:
// 			//stopper channel reads after 5 seconds
// 			log.Fatalf("timeout %s", t)
// 		}
// 	}
// }

// ?EXAMPLE 06 SELECT:
// 👀👀👀

// func main() {
// 	chans := []chan int{
// 		make(chan int),
// 		make(chan int),
// 	}

// 	for i := range chans {
// 		// starting 2 go-routines each one has its own channel
// 		go func(i int, ch chan<- int) {
// 			for {
// 				time.Sleep(time.Duration(i) * time.Second)
// 				ch <- i
// 			}
// 		}(i+1, chans[i])
// 	}

// 	oneCounter, twoCounter := 0, 0
// 	for i := 0; i < 12; i++ {
// 		// 👍 Pooling means waiting for input from multiple channels sequentially
// 		// The issue here is that if one channel has data available more frequently than the other, the loop might be unnecessarily delayed
// 		// i1 := <-chans[0]
// 		// fmt.Println("received", i1)
// 		// i2 := <-chans[1]
// 		// fmt.Println("received", i2)
// 		//👍 SELECT:- allows you to listen to multiple channels simultaneously and proceed with the one that becomes ready first

// 		select {
// 		case m0 := <-chans[0]:
// 			fmt.Println("received", m0)
// 			oneCounter++
// 		case m1 := <-chans[1]:
// 			fmt.Println("received", m1)
// 			twoCounter++
// 		}
// 	}
// 	fmt.Println(oneCounter, twoCounter)
// }

// ?EXAMPLE 05 PRIME SEIVE (LEAVE IT)

// func generator(limit int, ch chan<- int) {
// 	for i := 2; i < limit; i++ {
// 		ch <- i
// 	}

// 	close(ch) // generator closes
// }

// func filter(src <-chan int, dst chan<- int, prime int) {
// 	for i := range src {
// 		// When src channel closes, this loop is done
// 		// i is the integer value coming from the src channel
// 		if i%prime != 0 {
// 			// If it's not divisible by the prime no. pass it on
// 			dst <- i
// 		}
// 	}
// 	close(dst)
// }

// func sieve(limit int) {
// 	ch := make(chan int)

// 	go generator(limit, ch)

// 	for {
// 		prime, ok := <-ch

// 		if !ok {
// 			break
// 		}

// 		ch1 := make(chan int)
// 		go filter(ch, ch1, prime)

// 		ch = ch1

// 		fmt.Print(prime, "  ")
// 	}
// }

// func main() {
// 	sieve(100) // 2 3 5 7 11 13 17 19 ...
// 	fmt.Println()
// }

// ?EXAMPLE 04
// 👀👀👀

// var counter int

// func writer(ch chan<- int) {
// 	fmt.Println("START WRITING")
// 	ch <- counter
// 	counter++
// 	fmt.Println("END WRITING")
// }

// func reader(ch <-chan int) {
// 	fmt.Println("START READING")
// 	// val := <-c
// 	fmt.Println("END READING")
// }

// func main() {
// 	ch := make(chan int)
// 	go writer(ch)
// 	go reader(ch)
// 	fmt.Scanln()
// }

// ?EXAMPLE 03 GO WEB SERVER IS CONCURRENT! (LEAVE IT)

// type nextCh chan int

// // METHOD ON TYPE nextCh
// func (ch nextCh) handler(w http.ResponseWriter, r *http.Request) {
// 	fmt.Fprintf(w, "<h1>You got %d</h1>", <-ch)
// }

// func counter(ch chan<- int) {
// 	for i := 0; ; i++ {
// 		ch <- i
// 	}
// }

// func main() {
// 	var nextID nextCh = make(chan int)

// 	go counter(nextID)

// 	http.HandleFunc("/", nextID.handler)
// 	log.Fatal(http.ListenAndServe(":8080", nil))
// }

// ?EXAMPLE 02 GO WEB SERVER IS CONCURRENT!
// 👀👀👀 (IMPORTANT EXAMPLE)

// var nextID = make(chan int)

// func handler(w http.ResponseWriter, r *http.Request) {
// 	fmt.Fprintf(w, "<h1>You got %d</h1>", <-nextID) // Blocking call, can't read unless some go-routine writes to it
// }

// func counter() {
// 	for i := 0; ; i++ {
// 		nextID <- i // 👍 Blocking call unless a receiver reads  from the channel
// 	}
// }

// func main() {
// 	go counter()
// 	http.HandleFunc("/", handler)
// 	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
// 		fmt.Fprintf(w, "Favicon served")
// 	})
// 	log.Fatal(http.ListenAndServe(":8080", nil))
// }

// ?EXAMPLE 01 PARALLEL GET THROUGH HTTP
// 👀👀👀

// type result struct {
// 	url     string
// 	err     error
// 	latency time.Duration
// }

// func get(url string, ch chan<- result) {
// 	start := time.Now()

// 	if resp, err := http.Get(url); err != nil {
// 		ch <- result{url, err, 0} // 🌻writing to channel, blocking call until it finds some receiver
// 	} else {
// 		t := time.Since(start).Round(time.Millisecond)
// 		ch <- result{url, nil, t} // 🌻writing to channel, blocking call until it finds some receiver
// 		resp.Body.Close()
// 	}
// 	// fmt.Println("End")
// }

// func main() {
// 	results := make(chan result) // declaring channel of result type
// 	list := []string{
// 		"https://amazon.com",
// 		"https://google.com",
// 		"https://nytimes.com",
// 		"https://youtube.com",
// 	}

// 	for _, url := range list {
// 		go get(url, results)
// 	}

// 	for range list {
// 		// 🌻Reading from channel (blocking call, until some go-routine writes to the channel) , data flows in order
// 		r := <-results

// 		if r.err != nil {
// 			log.Printf("%-20s %s\n", r.url, r.err)
// 		} else {
// 			log.Printf("%-20s %s\n", r.url, r.latency)
// 		}
// 	}

// 	// time.Sleep(time.Millisecond * 3000) //blocking call
// 	//👍 DO NOT LOOP ON CHANNEL.

// 	// for i := range results {
// 	// r := <-results

// 	// if r.err != nil {
// 	// 	log.Printf("%-20s %s %s\n", r.url, r.latency, r.err)
// 	// } else {
// 	// 	log.Printf("%-20s %s\n", r.url, r.latency)
// 	// }
// 	// 	fmt.Println(i)
// 	// }
// }

// CONCURRENCY END
// ***********************************************************************************************************************
// ***********************************************************************************************************************
// ***********************************************************************************************************************
// ***********************************************************************************************************************
// ***********************************************************************************************************************
// ***********************************************************************************************************************

// GO OOPS START

// EXAMPLE 12
// WEB SERVER WITH CRUD ON HASH TABLE

// type dollars float32

// type database map[string]dollars

// func (d dollars) String() string {
// 	return fmt.Sprintf("$%.2f", d)
// }

// // add the handlers

// func (db database) list(w http.ResponseWriter, r *http.Request) {
// 	for item, price := range db {
// 		fmt.Fprintf(w, "%s: %v\n", item, price)
// 	}
// }

// func (db database) add(w http.ResponseWriter, r *http.Request) {
// 	item := r.URL.Query().Get("item")
// 	price := r.URL.Query().Get("price")

// 	if _, ok := db[item]; ok {
// 		msg := fmt.Sprintf("duplicate item: %q", item)
// 		http.Error(w, msg, http.StatusBadRequest) //400
// 		return
// 	}

// 	p, err := strconv.ParseFloat(price, 32)
// 	if err != nil {
// 		msg := fmt.Sprintf("invalid price: %q", price)
// 		http.Error(w, msg, http.StatusBadRequest) //400
// 		return
// 	}

// 	db[item] = dollars(p)
// 	fmt.Fprintf(w, "added %s with price %s\n", item, db[item])
// }

// func (db database) update(w http.ResponseWriter, r *http.Request) {
// 	item := r.URL.Query().Get("item")
// 	price := r.URL.Query().Get("price")

// 	if _, ok := db[item]; !ok {
// 		msg := fmt.Sprintf("no such item: %q", item)
// 		http.Error(w, msg, http.StatusNotFound) //404
// 		return
// 	}

// 	p, err := strconv.ParseFloat(price, 32)
// 	if err != nil {
// 		msg := fmt.Sprintf("invalid price: %q", price)
// 		http.Error(w, msg, http.StatusBadRequest) //400
// 		return
// 	}

// 	db[item] = dollars(p)
// 	fmt.Fprintf(w, "new price for item %s is %s\n", db[item], item)
// }

// func (db database) detail(w http.ResponseWriter, r *http.Request) {
// 	item := r.URL.Query().Get("item")

// 	if _, ok := db[item]; !ok {
// 		msg := fmt.Sprintf("no such item: %q", item)
// 		http.Error(w, msg, http.StatusNotFound) //404
// 		return
// 	}

// 	fmt.Fprintf(w, "item %s has price %s\n", db[item], item)
// }

// func (db database) delete(w http.ResponseWriter, r *http.Request) {
// 	item := r.URL.Query().Get("item")

// 	if _, ok := db[item]; !ok {
// 		msg := fmt.Sprintf("no such item: %q", item)
// 		http.Error(w, msg, http.StatusNotFound) //404
// 		return
// 	}

// 	delete(db, item)

// 	fmt.Fprintf(w, "deleted item %s\n", db[item])
// }

// func main() {
// 	db := database{
// 		"shoes": 50,
// 		"socks": 5,
// 	}

// 	// add the routes

// 	http.HandleFunc("/list", db.list)
// 	http.HandleFunc("/create", db.add)
// 	http.HandleFunc("/update", db.update)
// 	http.HandleFunc("/detail", db.detail)
// 	http.HandleFunc("/delete", db.delete)

// 	log.Fatal(http.ListenAndServe(":8080", nil))
// }

// EXAMPLE 11
// By using interfaces, you can write functions that operate on different types as long as they satisfy the required interface.
// This is powerful because it allows you to write more generic and reusable code.

// getArea function takes any Shape as a parameter. This function doesn't need to know whether it's
// dealing with a circle or a rectangle; it just cares that the shape has an Area method.

// type Circle struct {
// 	Radius float64
// }

// type Rectangle struct {
// 	Length, Width float64
// }

// type Shape interface {
// 	Area() float64
// }

// func (c Circle) Area() float64 {
// 	return (c.Radius * c.Radius) * 22 / 7
// }

// func (r Rectangle) Area() float64 {
// 	return (r.Length * r.Width)
// }

// func getArea(shape Shape) {
// 	fmt.Println(shape.Area())
// }

// func main() {
// 	r := Rectangle{4, 5}
// 	r2 := Rectangle{20, 2.5}
// 	c := Circle{7}

// 	shapes := []Shape{r, c, r2}

// 	for _, sh := range shapes {
// 		// fmt.Println(sh.Area())
// 		getArea(sh)
// 	}
// }

// EXAMPLE 10
// Method Values

// type Point struct {
// 	X, Y float64
// }

// func (p Point) Distance(q Point) float64 {
// 	return math.Hypot(p.X-q.X, p.Y-q.Y)
// }

// func main() {
// 	p1 := Point{0, 0}
// 	points := []Point{{1, 2}, {3, 4}, {5, 6}, {7, 8}, {1, 1}}
// 	distanceFromOrigin := p1.Distance
// 	for _, p := range points {
// 		fmt.Println(distanceFromOrigin(p))
// 	}
// }

// EXAMPLE 09
// CURRYING

// func Multiply(a, b int) int { return a * b }

// func MultiplyByN(n int) func(int) int {
// 	return func(b int) int {
// 		return Multiply(n, b)
// 	}
// }

// func main() {
// 	mulby10 := MultiplyByN(10)
// 	for i := 1; i <= 10; i++ {
// 		fmt.Println(mulby10(i))
// 	}
// }

// EXAMPLE 08

// sort.Sort(sort.interface) takes an interface with 3 methods:- Len(), Less(), Swap()
// Any Type which has all 3 methods can implement this interface

// type Organ struct {
// 	Name   string
// 	Weight int
// }

// type Organs []Organ

// func (s Organs) Len() int {
// 	return len(s)
// }

// func (s Organs) Swap(i, j int) {
// 	s[i], s[j] = s[j], s[i]
// }

// type ByName struct {
// 	Organs
// }

// type ByWeight struct {
// 	Organs
// }

// func (s ByName) Less(i, j int) bool {
// 	return s.Organs[i].Name < s.Organs[j].Name
// }

// func (s ByWeight) Less(i, j int) bool {
// 	return s.Organs[i].Weight < s.Organs[j].Weight
// }

// func main() {
// 	s := []Organ{{"Brain", 1200}, {"Allevolie", 12}, {"Heart", 280}, {"Pancreas", 350}, {"Liver", 3210}}
// 	sort.Sort(ByName{s})
// 	fmt.Println(s)
// 	sort.Sort(ByWeight{s})
// 	fmt.Println(s)
// }

// func main() {
// 	s := []Organ{{"Brain", 1200}, {"Heart", 280}, {"Pancreas", 350}, {"Liver", 3210}}
// 	sort.Slice(s, func(i, j int) bool {
// 		return s[i].Weight > s[j].Weight
// 	})
// 	fmt.Println(s)
// }

// EXAMPLE 07

// type Pair struct {
// 	Path string
// 	Hash string
// }

// type PairWithLength struct {
// 	Pair
// 	Length int
// }

// func (p Pair) String() string {
// 	return fmt.Sprintf("Hash of %s is %s", p.Path, p.Hash)
// }

// func (p PairWithLength) String() string {
// 	return fmt.Sprintf("Hash of %s is %s with Length %d", p.Path, p.Hash, p.Length)
// }

// func (p Pair) Filename() string {
// 	return filepath.Base(p.Path)
// }

// // Any type (Pair, PairWithLength) which has Filename() method can implement this interface
// type Filenamer interface {
// 	Filename() string
// }

// func main() {
// 	p := Pair{
// 		Path: "usr/lib",
// 		Hash: "0xde0d",
// 	}
// 	var fn Filenamer = PairWithLength{Pair{"usr/bin", "0xc2od"}, 120}

// 	fmt.Println(p.Filename())
// 	fmt.Println(fn.Filename())
// }

// EXAMPLE 06

// type Text string

// func (t *Text) Write(p []byte) (int, error) {
// 	*t = Text(string(p))
// 	return len(p), nil
// }

// func main() {
// 	var t Text
// 	if len(os.Args) > 1 {
// 		file, _ := os.Open(os.Args[1])
// 		io.Copy(&t, file)
// 	}
// 	fmt.Printf("%v\n", string(t))
// 	fmt.Printf("%d bytes written\n", len(t))
// }

// EXAMPLE 05

// type Point struct {
// 	X, Y float64
// }

// type Line struct {
// 	Begin, End Point
// }

// func (l Line) Distance() float64 {
// 	return math.Hypot(l.End.X-l.Begin.X, l.End.Y-l.Begin.Y)
// }

// func (l Line) ScaleBy(f float64) Line {
// 	l.End.X += (f - 1) * (l.End.X - l.Begin.X)
// 	l.End.Y += (f - 1) * (l.End.Y - l.Begin.Y)
// 	return Line{l.Begin, Point{l.End.X, l.End.Y}}
// }

// func main() {
// 	l1 := Line{Point{1, 2}, Point{4, 6}}
// 	l2 := l1.ScaleBy(2.5)
// 	fmt.Println(l1.Distance())
// 	fmt.Println(l2.Distance())
// 	// Method chaining, only possible as ScaleBy() is returning Line
// 	fmt.Println(Line{Point{1, 2}, Point{4, 6}}.ScaleBy(10).Distance())
// }

// EXAMPLE 04 (IMPORTANT!)

// type Point struct {
// 	X, Y float64
// }

// type Line struct {
// 	Begin, End Point
// }

// type Path []Point

// func (l Line) Distance() float64 {
// 	return math.Hypot(l.End.X-l.Begin.X, l.End.Y-l.Begin.Y)
// }

// func (p Path) Distance() (sum float64) {
// 	for i := 1; i < len(p); i++ {
// 		sum += Line{p[i-1], p[i]}.Distance()
// 	}
// 	return sum
// }

// type Distancer interface {
// 	Distance() float64
// }

// // Line and Path both UDT satisfy the Distancer interface as both have the Distance() method
// func PrintDistance(d Distancer) {
// 	fmt.Println(d.Distance())
// }

// func main() {
// 	l := Line{Point{1, 2}, Point{4, 6}}
// 	path := Path{{1, 1}, {5, 1}, {5, 4}, {1, 1}}
// 	PrintDistance(l)    // l.DIstance()
// 	PrintDistance(path) // path.Distance()
// }

// EXAMPLE 03 (IMPORTANT EXAMPLE)

// io.Writer is an interface with Write(p []byte) method
// io.Reader is an interface with Read(p []byte) method
// in io.Copy(dest, src), anything with the Read() method can be the src
// and anything with the Write() method can be the dest

// type ByteCounter int // concrete type, providing behaviour of a Writer

// func (b *ByteCounter) Write(p []byte) (int, error) {
// 	// I need a pointer to the receiver (ByteCounter) as i need to modify it
// 	l := len(p)
// 	*b += ByteCounter(l)
// 	return l, nil
// }

// func main() {
// 	var c ByteCounter
// 	f1, _ := os.Open("ip.txt")
// 	// f2, _ := os.Create("out.txt")
// 	f2 := &c

// 	n, _ := io.Copy(f2, f1) // io.Copy(dest io.Writer, interface src io.Reader interface)

// 	fmt.Println("copied", n, "bytes")
// 	fmt.Println(c)
// }

// EXAMPLE 02
//  A method is a function associated with a type
// Both Offset and Move are methods because they have a receiver of type Point
// and *Point, respectively. They operate on instances of the Point type.

// type Point struct {
// 	x float64
// 	y float64
// }

// func (p Point) Offset(x float64, y float64) Point {
// 	return Point{p.x + x, p.y + y}
// }

// func (p *Point) Move(x float64, y float64) {
// 	p.x += x
// 	p.y += y
// }

// func main() {
// 	p := Point{
// 		x: 0,
// 		y: 0,
// 	}
// 	fmt.Println(p)
// 	p1 := p.Offset(13, 13)
// 	fmt.Println(p1)
// 	fmt.Println(p)
// 	p.Move(4, 5)
// 	fmt.Println(p)
// }

// EXAMPLE 01

// type IntSlice []int

// func (is IntSlice) String() string {
// 	var strs []string

// 	for _, v := range is {
// 		strs = append(strs, strconv.Itoa(v))
// 	}
// 	return "[" + strings.Join(strs, ";") + "]"
// }

// func main() {
// 	var v IntSlice = []int{1, 2, 3}

// 	// s is an interface, what can I assign to an interface?
// 	// I can assign to it anything that satisfies the interface (any actual type that has a String method here!)
// 	var s fmt.Stringer = v

// 	for i, v := range v {
// 		fmt.Printf("v[%d]: %d\n", i, v)
// 	}
// 	fmt.Printf("%T %[1]v\n", v)
// 	fmt.Printf("%T %[1]v\n", s)
// }

// OOPS END

// *************************************************************************************************************************
// *************************************************************************************************************************
// *************************************************************************************************************************
// *************************************************************************************************************************
// *************************************************************************************************************************

// EXAMPLE 24
// Client Server

// type todo struct {
// 	UserID    int    `json:"userID"`
// 	ID        int    `json:"id"`
// 	Title     string `json:"title"`
// 	Completed bool   `json:"completed"`
// }

// var form = `
// <h1>Todo #{{.ID}}</h1>
// <div style="color:green">{{.UserID}}</div>
// <div>{{.Title}}</div>
// <h3>Completed {{.Completed}}</h3>
// `

// func handler(w http.ResponseWriter, r *http.Request) {
// 	const base = "https://jsonplaceholder.typicode.com/"
// 	resp, err := http.Get(base + r.URL.Path)

// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusServiceUnavailable)
// 		return
// 	}

// 	defer resp.Body.Close()

// 	// body, err := io.ReadAll(resp.Body)

// 	var item todo
// 	// err = json.Unmarshal(body, &item)

// 	if json.NewDecoder(resp.Body).Decode(&item); err != nil {
// 		http.Error(w, err.Error(), http.StatusServiceUnavailable)
// 		return
// 	}
// 	tmp1 := template.New("mine")
// 	tmp1.Parse(form)
// 	tmp1.Execute(w, item)
// }

// func main() {
// 	http.HandleFunc("/", handler)
// 	log.Fatal(http.ListenAndServe(":8080", nil))
// }

// EXAMPLE 23
// Http Server

// func handler(w http.ResponseWriter, r *http.Request) {
// 	fmt.Println("Body: ", r.Body)
// 	fmt.Println("Header: ", r.Header)
// 	fmt.Println("Form: ", r.Form)
// 	fmt.Println("Method: ", r.Method)
// 	fmt.Println("Host", r.Host)
// 	fmt.Println("Query", r.URL.Query())
// 	fmt.Println("Raw Query", r.URL.RawQuery)
// 	fmt.Fprintf(w, "Hello, World from %s\n", r.URL.Path)
// }

// func main() {
// 	http.HandleFunc("/", handler)
// 	log.Fatal(http.ListenAndServe(":8080", nil))
// }

// EXAMPLE 22
// REGEX

// func main() {
// 	text := "a aba abba abbba abbbba"
// 	re := regexp.MustCompile("b+")

// 	res := re.FindAllString(text, -1)
// 	for _, v := range res {
// 		fmt.Println(v)
// 	}
// 	fmt.Println(res)
// }

// EXAMPLE 21
// Slices gotcha & fix like closures

// func main() {
// 	a := [3]int{1, 2, 3}
// 	b := a[:1]
// 	c := a[:2]

// 	fmt.Printf("%d %d %v\n", len(a), cap(a), a)
// 	fmt.Printf("%d %d %v\n", len(b), cap(b), b)
// 	fmt.Printf("%d %d %v\n", len(c), cap(c), c)
// 	fmt.Println("------------------------------------------------------")
// 	c = append(c, 9)
// 	fmt.Printf("%d %d %v\n", len(a), cap(a), a)
// 	fmt.Printf("%d %d %v\n", len(b), cap(b), b)
// 	fmt.Printf("%d %d %v\n", len(c), cap(c), c)

// 	for i, _ := range a {
// 		fmt.Printf("%d %p\n", a[i], &a[i])
// 	}
// 	fmt.Println("***************************************")
// 	for i, _ := range b {
// 		fmt.Printf("%d %p\n", b, &b[i])
// 	}
// 	fmt.Println("***************************************")
// 	for i, _ := range c {
// 		fmt.Printf("%d %p\n", c[i], &c[i])
// 	}
// }

// func main() {
// 	a := [][2]int{{1, 2}, {3, 4}, {5, 6}}
// 	b := make([][]int, 0, 3)

// 	for _, item := range a {
// 		i := make([]int, len(item))
// 		copy(i, item[:]) // make unique
// 		b = append(b, i[:])
// 	}
// 	fmt.Println(a)
// 	fmt.Println(b)
// }

// EXAMPLE 21

// func fib() func() int {
// 	a, b := 0, 1

// 	return func() int {
// 		a, b = b, a+b
// 		return b
// 	}
// }

// func main() {
// 	f := fib()
// 	g := fib()
// 	fmt.Printf("%d %d %d\n", f(), f(), f())
// 	fmt.Printf("%d %d %d\n", g(), g(), g())
// 	fmt.Println(f(), g())
// }

// func print(arr []*int) {
// 	for i, _ := range arr {
// 		fmt.Printf("%d ", *arr[i])
// 	}
// 	fmt.Println()
// }

// func main() {
// 	a := []int{1, 2, 3}
// 	b := []*int{}

// 	for _, val := range a {
// 		b = append(b, &val)
// 	}
// 	fmt.Println(a)
// 	fmt.Println(b)
// 	print(b)
// }

// EXAMPLE 20

// Stale slice pointer problem

// func main() {
// 	a := []int{1}
// 	p := &a[0]
// 	// *(p) = -90
// 	fmt.Printf("%d %d %v %[3]p\n", len(a), cap(a), a)
// 	a = append(a, 2)
// 	*(p) = -90
// 	fmt.Printf("%d %d %v %[3]p\n", len(a), cap(a), a)
// }

// EXAMPLE 19
// JSON & Structs

// type Response struct {
// 	Page  int      `json:"page"`
// 	Words []string `json:"words,omitempty"`
// }

// func main() {
// 	r1 := Response{
// 		1,
// 		[]string{"solo", "mountains", "green-lands"},
// 	}
// 	j, err := json.Marshal(r1)
// 	if err != nil {
// 		fmt.Fprintln(os.Stderr, err)
// 		os.Exit(-1)
// 	}
// 	fmt.Println(string(j))

// 	var r2 Response

// 	err = json.Unmarshal(j, &r2)
// 	if err != nil {
// 		fmt.Fprintln(os.Stderr, err)
// 		os.Exit(-1)
// 	}
// 	fmt.Println(r1)
// 	fmt.Println(r2)
// }

// EXAMPLE 18

// type Album struct {
// 	title string
// }

// func changeTitle(album *Album) {
// 	album.title = "UNDEFINED"
// }

// func main() {
// 	a1 := Album{"The Green Fields"}
// 	fmt.Println(a1)
// 	changeTitle(&a1)
// 	fmt.Println(a1)
// }

// EXAMPLE 17

// VALID
// func main() {
// 	m := map[int]*int{}
// 	a := 10
// 	m[10] = &a
// 	fmt.Println(m[10])
// }

// EXAMPLE 16

// func main() {
// 	type Album struct {
// 		title string
// 	}

// 	a := Album{"The Black Album"}
// 	b := &Album{"The Green Album"}

// 	fmt.Println(a)
// 	fmt.Println(b)
// 	a.title = "Black Changed Changed!"
// 	b.title = "Hey Green Grass Changed"
// 	fmt.Println(a)
// 	fmt.Println(b)
// }

// EXAMPLE 15

// type Employee struct {
// 	Name  string
// 	Email string
// 	Age   int
// 	Boss  *Employee
// 	Hired time.Time
// }

// func main() {
// 	emps := map[string]*Employee{}

// 	emps["matt"] = &Employee{
// 		Name:  "matt",
// 		Email: "holiday@matt.com",
// 		Age:   53,
// 		Boss:  nil,
// 		Hired: time.Now(),
// 	}

// 	emps["ashiq"] = &Employee{
// 		Name:  "ashiq",
// 		Email: "a@a.com",
// 		Age:   24,
// 		Boss:  emps["matt"],
// 		Hired: time.Now(),
// 	}

// 	emps["ashiq"].Age += 10
// 	fmt.Println(*emps["matt"])
// 	fmt.Println(*emps["ashiq"])
// }

// EXAMPLE 14

// var raw = `
// <html>
// <body>
//     <h1>First Heading</h1>
//     <p>First Paragraph</p>
//     <p>HTML images are defined with the img tag:</p>
//     <img src="abc.jpg" width="100" height="100">
// </body>
// </html>
// `

// func visit(node *html.Node, words *int, pics *int) {

// 	fmt.Println("Node:-", node.Data)
// 	if node.Type == html.TextNode {
// 		(*words) += len(strings.Fields(node.Data))
// 	} else if node.Type == html.ElementNode && node.Data == "img" {
// 		(*pics)++
// 	}
// 	for c := node.FirstChild; c != nil; c = c.NextSibling {
// 		visit(c, words, pics)
// 	}
// }

// func countWordsAndImages(doc *html.Node) (int, int) {
// 	var words, pics int
// 	visit(doc, &words, &pics)
// 	return words, pics
// }

// func main() {
// 	doc, err := html.Parse(bytes.NewReader([]byte(raw)))

// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "parse failed: %s", err)
// 		os.Exit(-1)
// 	}
// 	words, pics := countWordsAndImages(doc)
// 	fmt.Printf("#words:%d & #pics:%d\n", words, pics)
// }

// EXAMPLE 13

// func main() {
// 	a := [3]int{1, 2, 3}

// 	b := a[:2]
// 	c := a[:2:2]
// 	fmt.Printf("A:%d %d %p %v\n", len(a), cap(a), &a, a)
// 	fmt.Printf("B:%d %d %p %v\n", len(b), cap(b), b, b)
// 	fmt.Printf("C:%d %d %p %v\n", len(c), cap(c), c, c)

// 	b = append(b, -10)
// 	a[0] = -78
// 	c = append(c, 99)
// 	fmt.Printf("A:%d %d %p %v\n", len(a), cap(a), &a, a)
// 	fmt.Printf("B:%d %d %p %v\n", len(b), cap(b), b, b)
// 	fmt.Printf("C:%d %d %p %v\n", len(c), cap(c), c, c)
// }

// EXAMPLE 12
// Closures

// func main() {
// 	s := make([]func(), 4)

// 	for i := 0; i < 4; i++ {
// 		i2 := i // closure capture
// 		s[i] = func() {
// 			fmt.Printf("%d %p\n", i2, &i2)
// 		}
// 	}

// 	for i := 0; i < 4; i++ {
// 		s[i]()
// 	}
// }

// func fib() func() int {
// 	a, b := 0, 1

// 	return func() int {
// 		a, b = b, a+b
// 		return b
// 	}
// }

// func main() {
// 	f, g := fib(), fib()
// 	fmt.Println(f(), f(), f(), f(), f())
// 	fmt.Println(g(), g(), g(), g(), g())
// }

// EXAMPLE 11
// defer GOTCHAS

// func main() {
// 	a := 10
// 	defer fmt.Println(&a)
// 	a = a * a * a * a
// 	fmt.Println(&a)
// }

// EXAMPLE 10

// func do(a1 *[2]int) {
// 	a1[0] = 99
// 	a1[1] = -100
// }

// func main() {
// 	a := [2]int{1, 2}
// 	fmt.Println(a)
// 	do(&a)
// 	fmt.Println(a)
// }

// func do(m1 *map[int]int) {
// 	(*m1)[9] = 99
// 	(*m1) = make(map[int]int)
// }

// func main() {
// 	m := map[int]int{
// 		1: 11,
// 		2: 22,
// 		3: 33,
// 	}
// 	fmt.Println(m)
// 	do(&m)
// 	fmt.Println(m)
// }

// func do(m1 map[int]int) {
// 	m1[9] = 99
// 	m1 = make(map[int]int)
// }

// func main() {
// 	m := map[int]int{
// 		1: 11,
// 		2: 22,
// 		3: 33,
// 		4: 44,
// 	}
// 	fmt.Println(m)
// 	do(m)
// 	fmt.Println(m)
// }

// EXAMPLE 09
// Line count, word count, char count for a file

// func main() {
// 	for _, fname := range os.Args[1:] {
// 		var lc, wc, cc int
// 		file, err := os.Open(fname)

// 		if err != nil {
// 			fmt.Fprintln(os.Stderr, err)
// 			continue
// 		}

// 		scan := bufio.NewScanner(file)

// 		for scan.Scan() {
// 			s := scan.Text()

// 			wc += len(strings.Fields(s))
// 			cc += len(s)
// 			lc++
// 		}
// 		fmt.Printf("%5d %5d %5d %s\n", lc, wc, cc, fname)
// 	}
// }

// EXAMPLE 09
// Concat file content to the stdout

// func main() {
// 	for _, fname := range os.Args[1:] {
// 		file, err := os.Open(fname)

// 		if err != nil {
// 			fmt.Fprintln(os.Stderr, err)
// 			continue
// 		}

// 		// if _, err := io.Copy(os.Stdout, file); err != nil {
// 		// 	fmt.Fprintln(os.Stderr, err)
// 		// 	continue
// 		// }

// 		data, err := io.ReadAll(file)

// 		if err != nil {
// 			fmt.Fprintln(os.Stderr, err)
// 			continue
// 		}

// 		fmt.Println(string(data)) // convert byte slice to string
// 		fmt.Println("The file has", len(data), "bytes")
// 		file.Close()
// 	}
// }

// EXAMPLE 08

// func main() {
// 	hashMap := make(map[string]int)
// 	names := []string{"ashiq", "amir", "zahid", "yawer", "bilal", "owais", "muzi"}

// 	for index, value := range names {
// 		hashMap[value] = index*index + index
// 	}
// 	for key, value := range hashMap {
// 		fmt.Printf("%v => %v\n", key, value)
// 	}
// }

// func main() {
// 	var arr [3]int
// 	for index, value := range arr {
// 		arr[index] = (value+1)*(value+2) + index*index
// 	}
// 	for index, value := range arr {
// 		fmt.Println(index, value)
// 	}
// }

// func main() {
// 	n := 10

// 	for i := 0; i < 4; i++ {
// 		for j := 0; j < 4; j++ {
// 			fmt.Printf("(%v, %v)\n", i, j)
// 		}
// 	}

// 	price, n := 4, n*10
// 	fmt.Println(n, price)
// }

// EXAMPLE 07
// program to count the top 3 words in a file

// func main() {
// 	scan := bufio.NewScanner(os.Stdin)
// 	words := make(map[string]int)

// 	scan.Split(bufio.ScanWords)

// 	for scan.Scan() {
// 		words[scan.Text()]++
// 	}
// 	fmt.Println("unique words in file:-", len(words))

// 	type kv struct {
// 		key string
// 		val int
// 	}

// 	var ss []kv

// 	for k, v := range words {
// 		ss = append(ss, kv{k, v})
// 	}

// 	sort.Slice(ss, func(i int, j int) bool {
// 		return ss[i].val > ss[j].val
// 	})
// 	for _, s := range ss[:3] {
// 		fmt.Println(s.key, "appears", s.val, "times")
// 	}
// }

// EXAMPLE 06
// i.ARRAYS ARE GETTING COPIED IN GO (Chunk of Memory Copied)
// ii.SLICES STRINGS and MAPS are refrenced (even when passing to functions also), their descriptor is going to be copied but the memory chunk they are
// pointing is getting refrenced (shared)

// func main() {
// 	a := [4]int{1, 2, 3, 4}
// 	b := a

// 	for i := 0; i < 4; i++ {
// 		b[i] += 10
// 	}
// 	fmt.Println(a, len(a))
// 	fmt.Println(b, len(b))
// }

// EXAMPLE 05

// func main() {
// 	if len(os.Args) < 3 {
// 		fmt.Fprintln(os.Stderr, "not enough args")
// 		os.Exit(-1)
// 	}

// 	old, new := os.Args[1], os.Args[2]
// 	scan := bufio.NewScanner(os.Stdin)

// 	for scan.Scan() {
// 		s := strings.Split(scan.Text(), old)
// 		t := strings.Join(s, new)

// 		fmt.Println(t)
// 	}
// }

// EXAMPLE 04
// GO STRINGS ARE UTF-8 ENCODING OF UNICODE CHARACTERS

// func main() {
// 	s := "ělite"
// 	fmt.Printf("%T %[1]v %d\n", s, len(s))
// 	fmt.Printf("%T %[1]v %d\n", []rune(s), len(s))
// 	fmt.Printf("%T %[1]v %d\n", []byte(s), len(s))
// }

// EXAMPLE 03

// func main() {
// 	var sum float64
// 	var n int

// 	for {
// 		var val float64

// 		_, err := fmt.Fscanln(os.Stdin, &val)
// 		if err != nil {
// 			fmt.Println(err)
// 			break
// 		}

// 		sum += val
// 		n++
// 	}

// 	if n == 0 {
// 		fmt.Fprintln(os.Stderr, "no values")
// 		os.Exit(-1)
// 	}

// 	fmt.Println("The average is", sum/float64(n))
// }

// EXAMPLE 02

// func main() {
// 	a := 10
// 	b := 3.4

// 	fmt.Printf("a: %8T %v\n", a, a)
// 	fmt.Printf("b: %8T %v\n", b, b)

// 	fmt.Printf("a: %8T %[1]v\n", a)
// 	fmt.Printf("b: %8T %[1]v\n", b)

// 	a = int(b)
// 	b = float64(a)

// 	fmt.Printf("a: %8T %[1]v\n", a)
// 	fmt.Printf("b: %8T %[1]v\n", b)
// }

// EXAMPLE 01

// func main() {
// 	fmt.Println("Hello World!")
// }
