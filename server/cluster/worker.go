package main

import (
	zmq "github.com/pebbe/zmq4"
	"log"

)

type WorkerRequest struct {
	data          []byte
	resultChannel chan []byte
}

type Worker struct {
	serviceAddress string
	requestChannel chan *WorkerRequest
}

func NewWorker(serviceAddress string) *Worker {
	w := &Worker{
		serviceAddress: serviceAddress,
		requestChannel: make(chan *WorkerRequest),
	}
	go func() {
		w.start()
	}()

	return w
}

func (w *Worker) DoRequest(data []byte) chan []byte {
	//log.Printf("Worker[]%s got request", w.serviceAddress)
	req := &WorkerRequest{
		data:          data,
		resultChannel: make(chan []byte),
	}
	w.requestChannel <- req
	return req.resultChannel
}

func (w *Worker) start() {
	requester, _ := zmq.NewSocket(zmq.REQ)
	defer requester.Close()
	err := requester.Connect(w.serviceAddress)
	if err != nil {log.Fatal(err)}
	log.Printf("Worker %s ready\n", w.serviceAddress)
	for wreq := range w.requestChannel {
		//log.Printf("Sending request %d bytes", len(clientRequest.requestData))
		requester.SendBytes(wreq.data, 0)

		//log.Println("Waiting response..")

		replyBytes, err := requester.RecvBytes(0)
		//log.Println("Got response..")
		if err != nil {log.Println(err)}
		//log.Println(replyBytes)
		wreq.resultChannel <- replyBytes
		close(wreq.resultChannel)
	}


}