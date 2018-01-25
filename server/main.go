package main

import (
	zmq "github.com/pebbe/zmq4"
	"log"
	"github.com/golang/protobuf/proto"
	"cyclops/proto"
	"fmt"
	"flag"
)

//var PORT = 5555
//var HASH_LIMIT = 50000000

func processMsg(msg []byte) (response []byte, err error) {

	request := &model.Request{}
	err = proto.Unmarshal(msg, request); if err != nil {return}

	switch request.Type {
	case model.RequestType_ADD_HASH:
		cmd := &model.AddHash{}
		err = proto.Unmarshal(request.Payload, cmd); if err != nil {return}
		log.Printf("Adding %d hahses", len(cmd.HashSet.Hashes))
		for _, hash := range cmd.HashSet.Hashes {
			addHash(hash)
			//log.Printf("Hash added %s", hash)
		}

		break
	case model.RequestType_SEARCH:
		cmd := &model.Search{}
		err = proto.Unmarshal(request.Payload, cmd); if err != nil {return}
		//log.Printf("Search H:%d, D:%d", cmd.Hash, cmd.Distance)
		results := search(cmd.Hash, uint64(cmd.Distance))
		response, err = proto.Marshal(&model.SearchResults{
			Hashes: results,
		})
		if err != nil {log.Println(err)}
		//log.Println(response)
		break
	}

	return
}


func main() {

	hashLimit := flag.Uint64("limit", 1000000, "Maximum hash limit for this process")
	port := flag.Uint("port", 5555, "TCP Service Port")


	flag.Parse()

	log.Printf("Hash limit set to %d", *hashLimit)
	initializeStorage(*hashLimit)

	responder, err := zmq.NewSocket(zmq.REP)
	if err != nil {log.Fatal(err)}
	defer responder.Close()

	tcpAddress := fmt.Sprintf("tcp://*:%v", *port)
	responder.Bind(tcpAddress)
	log.Printf("Service started at %s", tcpAddress)

	for {
		msg, err := responder.RecvBytes(0)
		if err != nil {
			log.Println("Message receive error ", err)
			continue
		}
		//log.Println("Message received.")
		responsePayload, err := processMsg(msg)
		response := &model.Response{}

		if err != nil {
			log.Println("Response failed. ", err)
			response.Type = model.ResponseType_ERROR
			response.Payload, _ = proto.Marshal(&model.ErrorInfo{
				Info: err.Error(),
			})
		} else {
			response.Type = model.ResponseType_SUCCESS
			response.Payload = responsePayload
		}
		responseBytes, _ := proto.Marshal(response)

		responder.SendBytes(responseBytes, 0)
	}
}
