package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type boolgen struct {
	src       rand.Source
	cache     int64
	remaining int
}

var defaultUpperLimit = 0x100
var dec, hex, mix = "d2x", "x2d", "both"
var averageTime int64
var iterations int64
var userSucceeded bool
var responseTime int64

func welcome() (int, string) {
	welcomeBanner := `
	__________
	| ________ |
	||12345678||
	|''''''''''|
	|[M|#|C][-]| HexDec - Become a Hexa(decimal) Pro!
	|[7|8|9][+]| Author: Ophir Harpaz (@ophirharpaz)
	|[4|5|6][x]| ascii art by hjw
	|[1|2|3][%]| 
	|[.|O|:][=]| Translated to GoLang by Quentin Rhoads-Herrera (@paragonsec)
	|==========|
    
	Set the game properties using:
	- the maximal number that will show up;
	- the game mode (d2x for decimal to hexa, x2d for the opposite direction, or both)
	`
	fmt.Println(welcomeBanner)

	//Taking user input for the max number
	var maxNumber int
	fmt.Printf("Choose a maximal number [default is %d]: ", defaultUpperLimit)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	input = strings.TrimSpace(input)
	convNum, err := strconv.Atoi(input)
	if err != nil {
		log.Fatal(err)
	}
	if convNum == 0 {
		maxNumber = defaultUpperLimit
	} else if convNum > 256 {
		fmt.Println("Number can't be greater than 256! Setting it to 256 for you :)")
		maxNumber = defaultUpperLimit
	} else {
		maxNumber = convNum
	}

	//Taking user input for game mode
	var gameMode string
	fmt.Printf("Game mode (x2d, d2x, both) [x2d]: ")
	input, err = reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	input = strings.TrimSpace(input)
	if input == "" {
		gameMode = "x2d"
	} else {
		gameMode = input
	}
	return maxNumber, gameMode
}

func (b *boolgen) Bool() bool {
	if b.remaining == 0 {
		b.cache, b.remaining = b.src.Int63(), 63
	}

	result := b.cache&0x01 == 1
	b.cache >>= 1
	b.remaining--

	return result
}

func new() *boolgen {
	return &boolgen{src: rand.NewSource(time.Now().UnixNano())}
}

func play(maxNumber int, gameMode string) {

	fmt.Println("To stop the game use CTRL+C")
	signalHandler()

	for true {
		switch {
		case mix == gameMode:
			r := new()
			r.Bool()
			userSucceeded, responseTime = playIterations(maxNumber, r.Bool())
		case dec == gameMode:
			userSucceeded, responseTime = playIterations(maxNumber, false)
		case hex == gameMode:
			userSucceeded, responseTime = playIterations(maxNumber, true)
		}
		iterations++
		averageTime = (averageTime*(iterations-1) + responseTime) / iterations
		if userSucceeded == false {
			fmt.Println("Oops, you got that last one wrong...")
			goodbye()
		}
	}
}

func randomInt(min, max int) int {
	return min + rand.Intn(max-min)
}

func playIterations(top int, hexTodec bool) (bool, int64) {
	rand.Seed(time.Now().UnixNano())
	n := randomInt(0, top)
	var nForDisplay string
	t := time.Now()
	secs := t.Unix()

	if hexTodec == true {
		nForDisplay = fmt.Sprintf("0x%x", n)
	} else {
		nForDisplay = strconv.Itoa(n)
	}

	var responseTime int64
	var convert string
	for true {
		fmt.Printf("* %s = ", nForDisplay)
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSuffix(input, "\r\n")
		t2 := time.Now()
		secs2 := t2.Unix()
		responseTime = secs2 - secs
		if hexTodec == true {
			newNum, _ := strconv.Atoi(input)
			convert = strconv.FormatInt(int64(newNum), 16)
			convert = "0x" + convert
		} else {
			input = "0x" + input
			newNum, _ := strconv.ParseInt(input, 0, 64)
			convert = strconv.FormatInt(int64(newNum), 10)
		}
		break
	}
	return convert == nForDisplay, responseTime
}

func goodbye() {
	fmt.Printf("\nYou placed %d iterations and your average response time was %d seconds.", iterations, averageTime)
	os.Exit(0)
}

func signalHandler() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\n CTRL+C pressed in terminal")
		goodbye()
	}()
}

func main() {
	maxNumber, gameMode := welcome()
	play(maxNumber, gameMode)

}
