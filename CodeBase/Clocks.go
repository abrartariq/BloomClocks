package main

import (
	"fmt"
	"hash/crc32"
	"hash/fnv"
	"hash/maphash"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/spaolacci/murmur3"
)

// TypeEVENT is MSG typed EVENT
var TypeEVENT int = 1

// TypePROCESS is MSG typed PROCESS
var TypePROCESS int = 2

// EventINT is INTERNAL
var EventINT int = -1

// EventSEND is SEND
var EventSEND int = -2

// EventRECV is RECV
var EventRECV int = -3

// EventEND is END
var EventEND int = -4

type event struct {
	etype        int
	pIDSend      int
	pIDRecv      int
	vClock       []int
	bClock       []int
	ePair        []int
	hasedIndices []int
}

type newprocess struct {
	pID   int
	pChan chan message
}

type message struct {
	msgType int
	mEvent  event
	pInfo   newprocess
}

// EMPTYEVENT Global For Empty Case
var EMPTYEVENT = event{0, 0, 0, nil, nil, nil, nil}

// EMPTYPROCESS Global For Empty Case
var EMPTYPROCESS = newprocess{0, nil}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	cmdArgs := os.Args

	first := time.Now()

	arg1, _ := strconv.ParseFloat(cmdArgs[1], 64)
	arg2, _ := strconv.ParseFloat(cmdArgs[2], 64)
	arg3, _ := strconv.ParseFloat(cmdArgs[3], 64)

	totalProcesses := int(arg1)
	bloomSize := int(arg2 * arg1)
	hashCount := int(arg3)

	hypervisor(totalProcesses, bloomSize, hashCount)

	last := time.Now()
	timeToPassive := last.Sub(first).Seconds()
	fmt.Println("TOTAL SECONDS TAKEN : ", timeToPassive)
}

func myhash(pair string, hashfunctions int, bloomSize int) []int {
	pairBytes := []byte(pair)
	switch hashfunctions {
	case 2:
		ansSlice := make([]int, 2)
		ansSlice[0] = int(mymumumr3(pairBytes) % uint32(bloomSize))
		ansSlice[1] = int(myfnv(pairBytes) % uint32(bloomSize))
		return ansSlice
	case 3:
		ansSlice := make([]int, 3)
		ansSlice[0] = int(mymumumr3(pairBytes) % uint32(bloomSize))
		ansSlice[1] = int(myfnv(pairBytes) % uint32(bloomSize))
		ansSlice[2] = int(mycrc32(pairBytes) % uint32(bloomSize))
		return ansSlice

	case 4:
		ansSlice := make([]int, 4)
		ansSlice[0] = int(mymumumr3(pairBytes) % uint32(bloomSize))
		ansSlice[1] = int(myfnv(pairBytes) % uint32(bloomSize))
		ansSlice[2] = int(mycrc32(pairBytes) % uint32(bloomSize))
		ansSlice[3] = int(mymaphash(pairBytes) % uint32(bloomSize))
		return ansSlice

	}
	return nil
}

func mymumumr3(pairBytes []byte) uint32 {
	hashedmurmur := murmur3.Sum32(pairBytes)
	return hashedmurmur
}

func myfnv(pairBytes []byte) uint32 {
	fnvHash := fnv.New32a()
	fnvHash.Write(pairBytes)
	fnvHashed := fnvHash.Sum32()
	return fnvHashed
}

func mycrc32(pairBytes []byte) uint32 {
	crc32Hashed := crc32.ChecksumIEEE(pairBytes)
	return crc32Hashed
}

func mymaphash(pairBytes []byte) uint32 {
	var mapHash maphash.Hash
	mapHash.Write(pairBytes)
	mapHashed := uint32(mapHash.Sum64())
	return mapHashed
}

func hypervisor(processcount int, bloomSize int, hashCount int) {

	endChans := make([]chan int, processcount)
	pMsgChans := make([]chan message, processcount)

	syncChan := make(chan bool)
	loggerEndChan := make(chan bool)
	loggerEventChan := make(chan event)

	currentPCount := 0

	go seqlogger(loggerEndChan, loggerEventChan)

	for i := 0; i < processcount; i++ {
		iCaptured := i
		procChan := make(chan message, processcount-1)
		endChan := make(chan int)

		pMsgChans[iCaptured] = procChan
		endChans[iCaptured] = endChan

		go createprocess(iCaptured, processcount, bloomSize, hashCount, procChan, endChan, loggerEventChan, syncChan)

		for j := 0; j < currentPCount; j++ {
			nProcsss := newprocess{iCaptured, procChan}
			var tempMsg = message{TypePROCESS, EMPTYEVENT, nProcsss}
			pMsgChans[j] <- tempMsg
		}

		for j := 0; j < currentPCount; j++ {
			nProcsss := newprocess{j, pMsgChans[j]}
			var tempMsg = message{TypePROCESS, EMPTYEVENT, nProcsss}
			procChan <- tempMsg
		}

		currentPCount++
	}

	for i := 0; i < processcount; i++ {
		<-syncChan
	}

	endroutines(endChans, processcount, loggerEndChan)
}

func endroutines(endChans []chan int, total int, loggerEndChan chan bool) {

	time.Sleep(time.Duration((total/100)*10) + 20*time.Second)

	start := time.Now()
	for i := 0; i < total; i++ {
		endChans[i] <- 0
	}
	end := time.Now()

	timeToPassive := (end.Sub(start)).Seconds() / 2
	time.Sleep(time.Duration(timeToPassive) * time.Second)

	for i := 0; i < total; i++ {
		endChans[i] <- 0
	}

	time.Sleep(5 * time.Second)

	loggerEndChan <- true
	<-loggerEndChan
}

func seqlogger(endChan chan bool, log chan event) {
	totalEvents := 0
	for {
		select {

		case val := <-log:
			if val.etype == EventINT {
				totalEvents++
				fmt.Print(totalEvents, ",", "INTERNAL EVENT", ",", val.pIDSend, ",", "-", ",", val.ePair, ",", val.hasedIndices, "\n")
				fmt.Println(val.vClock)
				fmt.Println(val.bClock)
			} else if val.etype == EventSEND {
				totalEvents++
				fmt.Print(totalEvents, ",", "SEND EVENT", ",", val.pIDSend, ",", val.pIDRecv, ",", val.ePair, ",", val.hasedIndices, "\n")
				fmt.Println(val.vClock)
				fmt.Println(val.bClock)

			} else if val.etype == EventRECV {
				totalEvents++
				fmt.Print(totalEvents, ",", "RECV EVENT", ",", val.pIDSend, ",", val.pIDRecv, ",", val.ePair, ",", val.hasedIndices, "\n")
				fmt.Println(val.vClock)
				fmt.Println(val.bClock)

			} else if val.etype == EventEND {
				fmt.Println("ENDING ", val.pIDSend)
				fmt.Println(val.vClock)
				fmt.Println(val.bClock)
			}
		case val := <-endChan:
			fmt.Println("TOTAL EVENTS OCCURED ", totalEvents, val)
			endChan <- true
			return
		}
	}
}

func createprocess(pid int, total int, bloomSize int, hashCount int, procChan chan message, endChan chan int, log chan event, sync chan bool) {
	pCount := 1
	vClock := make([]int, total)
	bClock := make([]int, bloomSize)
	myid := pid

	// SET CLOCK TO DEFAULT VALUE
	for i := 0; i < total; i++ {
		vClock[i] = 0
	}

	for i := 0; i < bloomSize; i++ {
		bClock[i] = 0
	}

	chans := make([]chan message, total)

	for {
		select {

		case val := <-procChan:
			if val.msgType == TypeEVENT {
				// Update V AND B CLOCK
				for i := 0; i < total; i++ {
					vClock[i] = max(vClock[i], val.mEvent.vClock[i])
				}

				// Update Local TICK
				vClock[myid]++

				for i := 0; i < bloomSize; i++ {
					bClock[i] = max(bClock[i], val.mEvent.bClock[i])
				}

				iPair := []int{myid, vClock[myid]}
				sPair := strconv.Itoa(myid) + "," + strconv.Itoa(vClock[myid])
				hasedIndices := myhash(sPair, hashCount, bloomSize)

				for _, val := range hasedIndices {
					bClock[val]++
				}

				var tempEvent = event{EventRECV, val.mEvent.pIDSend, myid, vClock, bClock, iPair, hasedIndices}
				log <- tempEvent

			} else if val.msgType == TypePROCESS {
				// update chan arrays
				pCount++
				if pCount == total {
					sync <- true
				}
				chans[val.pInfo.pID] = val.pInfo.pChan
			}

		case _ = <-endChan:
			var tempEvent = event{EventEND, myid, -99, vClock, bClock, nil, nil}
			pCount++
			if pCount == total+2 {
				log <- tempEvent
				return
			}

		default:
			if pCount == total {

				time.Sleep(100 * time.Millisecond)
				probValue := rand.Intn(100)
				probHandle := 0
				if probValue >= probHandle {
					// Send Event

					vClock[myid]++
					recvID := rand.Intn(total)
					for recvID == myid {
						recvID = rand.Intn(total)
					}

					iPair := []int{myid, vClock[myid]}
					sPair := strconv.Itoa(myid) + "," + strconv.Itoa(vClock[myid])
					hasedIndices := myhash(sPair, hashCount, bloomSize)

					for _, val := range hasedIndices {
						bClock[val]++
					}

					var tempEvent = event{EventSEND, myid, recvID, vClock, bClock, iPair, hasedIndices}
					var tempMsg = message{TypeEVENT, tempEvent, EMPTYPROCESS}

					log <- tempEvent
					chans[tempEvent.pIDRecv] <- tempMsg

				} else if probValue < probHandle {
					// Internal Event

					vClock[myid]++

					iPair := []int{myid, vClock[myid]}
					sPair := strconv.Itoa(myid) + "," + strconv.Itoa(vClock[myid])
					hasedIndices := myhash(sPair, hashCount, bloomSize)

					for _, val := range hasedIndices {
						bClock[val]++
					}

					var tempEvent = event{EventINT, myid, -99, vClock, bClock, iPair, hasedIndices}
					log <- tempEvent

				}
			}
		}
	}
}

func max(x, y int) int {
	if x < y {
		return y
	}
	return x
}
