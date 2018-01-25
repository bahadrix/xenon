package main

import (
	zmq "github.com/pebbe/zmq4"
	"fmt"
	"log"
	"cyclops/proto"
	"github.com/golang/protobuf/proto"
	"io/ioutil"
	"time"
)

var MAIN_SERVICE_ADDRESS = "tcp://127.0.0.1:6550"
var HTTP_PORT = 8076
var requestChannel = make(chan *ClientRequest)

type ClientRequest struct {
	requestData []byte
	responseChannel chan *ServiceReply
}

type ServiceReply struct {
	responseData []byte
	queryTime    time.Duration
	err          error
}

func DoRequest(requestData []byte) chan *ServiceReply{

	r := &ClientRequest{
		requestData:     requestData,
		responseChannel: make(chan *ServiceReply),
	}
	requestChannel <- r
	return r.responseChannel
}

func RequestSearch(hash uint64, distance uint32) (hashes []uint64, elapsed time.Duration, err error){
	s, _ := proto.Marshal(&model.Search{
		Hash:     hash,
		Distance: distance,
	})

	req, _ := proto.Marshal(&model.Request{
		Type:    model.RequestType_SEARCH,
		Payload: s,
	})
	reply := <-DoRequest(req)

	if reply.err != nil {
		err = reply.err
		return
	}

	elapsed = reply.queryTime

	response := &model.Response{}

	proto.Unmarshal(reply.responseData, response)

	if response.Type == model.ResponseType_SUCCESS {
		results := &model.SearchResults{}
		proto.Unmarshal(response.Payload, results)
		hashes = results.Hashes
	} else {
		errorInfo := &model.ErrorInfo{}
		proto.Unmarshal(response.Payload, errorInfo)
		err = fmt.Errorf(errorInfo.Info)
	}
	return
}

func RequestAddHash(hashSet []uint64) error {
	ah, _ := proto.Marshal(&model.AddHash{
		HashSet: &model.HashSet{
			Hashes: hashSet,
		},
	})

	req, err := proto.Marshal(&model.Request{
		Type:    model.RequestType_ADD_HASH,
		Payload: ah,
	})

	reply := <- DoRequest(req)
	if err != nil {return err}

	if reply.err == nil {
		response := &model.Response{}
		proto.Unmarshal(reply.responseData, response)
		if response.Type == model.ResponseType_ERROR{
			errorInfo := &model.ErrorInfo{}
			proto.Unmarshal(response.Payload, errorInfo)
			return fmt.Errorf(errorInfo.Info)
		}
	} else {
		return fmt.Errorf("Request error on loading hashes.")
	}
	return nil
}

func startClientLoop() {

	//  Socket to talk to server
	fmt.Printf("Connecting to service at %s\n", MAIN_SERVICE_ADDRESS)
	requester, _ := zmq.NewSocket(zmq.REQ)
	defer requester.Close()
	err := requester.Connect(MAIN_SERVICE_ADDRESS)
	if err != nil {log.Fatal(err)}

	for clientRequest := range requestChannel {
		//log.Printf("Sending request %d bytes", len(clientRequest.requestData))
		requester.SendBytes(clientRequest.requestData, 0)

		//log.Println("Waiting response..")

		start := time.Now()
		replyBytes, err := requester.RecvBytes(0)
		elapsed := time.Since(start)
		clientRequest.responseChannel <- &ServiceReply{
			responseData: replyBytes,
			err:          err,
			queryTime:    elapsed,
		}
		close(clientRequest.responseChannel)
	}
}

func loadHashesFromFile(filepath string, limit uint64 ) {
	log.Printf("Loading hashes from %s", filepath)
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Fatal(err)
	}

	hashSet := &model.HashSet{}
	err = proto.Unmarshal(data, hashSet)

	if limit != 0 {
		hashSet.Hashes = hashSet.Hashes[:limit]
	}

	log.Printf("First %d hash selected.", len(hashSet.Hashes))

	ah, _ := proto.Marshal(&model.AddHash{
		HashSet: hashSet,
	})

	req, err := proto.Marshal(&model.Request{
		Type:    model.RequestType_ADD_HASH,
		Payload: ah,
	})


	reply := <- DoRequest(req)

	if reply.err == nil {
		response := &model.Response{}
		proto.Unmarshal(reply.responseData, response)
		if response.Type == model.ResponseType_SUCCESS {
			log.Printf("Hashes successfully loaded.")
		} else {
			log.Printf("Error on loading hashes.")
		}
	} else {
		log.Printf("Request error on loading hashes.")
	}
}

func loadHashes(hashes []uint64) {

	ah, _ := proto.Marshal(&model.AddHash{
		HashSet: &model.HashSet{
			Hashes: hashes,
		},
	})

	req, _ := proto.Marshal(&model.Request{
		Type:    model.RequestType_ADD_HASH,
		Payload: ah,
	})

	reply := <-DoRequest(req)

	if reply.err == nil {
		response := &model.Response{}
		proto.Unmarshal(reply.responseData, response)
		if response.Type == model.ResponseType_SUCCESS {
			//log.Printf("Hashes successfully loaded.")
		} else {
			log.Printf("Error on loading hashes.")
		}
	} else {
		log.Printf("Request error on loading hashes.")
	}



}

func loadHashesByPart(filepath string) {
	log.Printf("Loading hashes from %s", filepath)
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Fatal(err)
	}

	hashSet := &model.HashSet{}
	err = proto.Unmarshal(data, hashSet)


	loadHashes(hashSet.Hashes[0:8000000])
	loadHashes(hashSet.Hashes[8000000:16000000])
	loadHashes(hashSet.Hashes[16000000:24000000])
	loadHashes(hashSet.Hashes[24000000:32000000])
}

func searchTest() {
	for i := 0; i < 10; i++ {
		go func(k int) {
			s, _ := proto.Marshal(&model.Search{
				Hash:     9223372036854775808,
				Distance: 0,
			})

			req, _ := proto.Marshal(&model.Request{
				Type:    model.RequestType_SEARCH,
				Payload: s,
			})
			log.Printf("Sending request")
			reply := <-DoRequest(req)

			if reply.err != nil {
				log.Fatal(reply.err)
			}

			response := &model.Response{}

			proto.Unmarshal(reply.responseData, response)

			if response.Type == model.ResponseType_SUCCESS {
				results := &model.SearchResults{}
				proto.Unmarshal(response.Payload, results)
				for _, h := range results.Hashes {
					fmt.Printf("(%d)", k)
					fmt.Printf("%b", h)
					fmt.Printf(" %v Elapsed (%s)\n", h, reply.queryTime)
				}
			}
		}(i)
	}
}

func main() {

	go func() {
		//loadHashesByPart("res/waplog_hashes.dat")
		//loadHashesByPart("res/waplog_hashes.dat")
		loadHashesFromFile("res/waplog_hashes.dat", 30000000)
		loadHashesFromFile("res/waplog_hashes.dat", 30000000)


		log.Printf("Hashes loaded.")
	}()


	go func() {
		//log.Printf("Starting HTTP server at %d", HTTP_PORT)
		getRouter().Run(fmt.Sprint(":",HTTP_PORT))
	}()

	startClientLoop()

}