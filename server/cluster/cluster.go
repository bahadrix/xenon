package main

import (
	zmq "github.com/pebbe/zmq4"
	"log"
	"fmt"
	"flag"
	"cyclops/proto"
	"github.com/golang/protobuf/proto"
	"sync"
)

var workers []*Worker
var rrIndex = 0

func createAddHashRequestBytes(hashes []uint64) []byte{
	wah := &model.AddHash{
		HashSet: &model.HashSet{
			Hashes: hashes,
		},
	}
	wahBytes, _ := proto.Marshal(wah)

	r := &model.Request{
		Type:    model.RequestType_ADD_HASH,
		Payload: wahBytes,
	}

	rBytes, _ := proto.Marshal(r)

	return rBytes
}

// Send round robin
func publishAddhash(msg []byte) []byte{

	request := &model.Request{}
	proto.Unmarshal(msg, request)
	ah := &model.AddHash{}
	proto.Unmarshal(request.Payload, ah)

	hashCount := len(ah.HashSet.Hashes)

	if hashCount == 1 {

		w := workers[rrIndex]
		rrIndex++
		if rrIndex >= len(workers) {
			rrIndex = 0
		}
		res := <-w.DoRequest(msg)
		//log.Printf("Hashes added to %d", rrIndex - 1)
		return res
	}

	// Add batch
	workerCount := len(workers)
	buckets := make([][]uint64, workerCount)
	residues := make([][]uint64, workerCount)

	buckSize := hashCount/workerCount

	if buckSize > 1 {
		for i := 0; i < workerCount; i++ {
			buckets[i] = ah.HashSet.Hashes[i*buckSize:(i+1)*buckSize]
		}
	}

	residue := hashCount % workerCount
	lastAddIndex := buckSize * workerCount
	for i := 0; i < residue; i++ {
		residues[rrIndex] = append(residues[rrIndex], ah.HashSet.Hashes[lastAddIndex + i])
		//log.Println("RR: %d V: %d", rrIndex, ah.HashSet.Hashes[lastAddIndex + i])
		rrIndex += 1
		if rrIndex >= workerCount {
			rrIndex = 0
		}
	}

	// Add buckets
	var wg sync.WaitGroup
	wg.Add(workerCount)
	resultChan := make(chan []byte, workerCount *2)
	for i := 0; i < workerCount; i++ {
		go func(buckIndex int) {
			defer wg.Done()
			w := workers[buckIndex]
			rBytes := createAddHashRequestBytes(buckets[buckIndex])
			resultBytes := <- w.DoRequest(rBytes)
			resultChan <- resultBytes

			if len(residues[buckIndex]) > 0 {
				rBytes = createAddHashRequestBytes(residues[buckIndex])
				resultBytes := <- w.DoRequest(rBytes)
				resultChan <- resultBytes
			}
		}(i)
	}
	wg.Wait()
	log.Println("Bulk done")
	close(resultChan)
	firstResult := <- resultChan

	return firstResult

}

// Send to all workers and collect results
func publishSearch(msg []byte) []byte {
	channels := make([]chan []byte, len(workers))
	i := 0
	for _, w := range workers {
		//log.Println("Waiting godot")
		channels[i] = w.DoRequest(msg)
		i++
	}
	searchResults := &model.SearchResults{
		Hashes: []uint64{},
	}

	for _, c := range channels {

		resultBytes := <- c

		resp := &model.Response{}
		proto.Unmarshal(resultBytes, resp)
		workerResult := &model.SearchResults{}
		err := proto.Unmarshal(resp.Payload, workerResult)
		if err != nil {log.Fatal(err)}

		searchResults.Hashes = append(searchResults.Hashes, workerResult.Hashes...)
	}

	resultBytes, _ := proto.Marshal(searchResults)
	respBytes, _ := proto.Marshal(&model.Response{
		Type:    model.ResponseType_SUCCESS,
		Payload: resultBytes,
	})

	return respBytes
}

func publishMessage(msg []byte) []byte {

	request := &model.Request{}
	err := proto.Unmarshal(msg, request)

	if err != nil { log.Println(err) }

	//log.Println(request)

	switch request.Type {
	case model.RequestType_ADD_HASH:
		return publishAddhash(msg)
	case model.RequestType_SEARCH:
		return publishSearch(msg)
	default:
		log.Fatal("Unknown request tye")
	}

	return nil
}

func main() {

	workers = []*Worker{
		NewWorker("tcp://localhost:6551"),
		NewWorker("tcp://localhost:6552"),
		NewWorker("tcp://localhost:6553"),
		NewWorker("tcp://localhost:6554"),
		NewWorker("tcp://localhost:6555"),
		NewWorker("tcp://localhost:6556"),
	}

	port := flag.Uint("port", 6550, "TCP Service Port")
	flag.Parse()

	responder, err := zmq.NewSocket(zmq.REP)
	if err != nil {log.Fatal(err)}
	defer responder.Close()

	tcpAddress := fmt.Sprintf("tcp://*:%v", *port)
	responder.Bind(tcpAddress)
	log.Printf("Service started at %s", tcpAddress)

	for {
		msg, err := responder.RecvBytes(0)
		if err != nil {log.Println(err)}
		response := publishMessage(msg)
		responder.SendBytes(response, 0)
	}
}