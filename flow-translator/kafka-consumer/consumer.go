/*
 * Copyright (c) 2018 Juniper Networks, Inc. All rights reserved.
 *
 * file:    main.go
 * details: Kafka Consumer based on confluent-kafka-go library
 *
 */
package kafkaconsumer

import (
	"os"
	"os/signal"

	msghandler "github.com/Juniper/collector/flow-translator/msg-handler"
	opts "github.com/Juniper/collector/flow-translator/options"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type consChannel struct {
	chConsName string
	inCh       chan []byte
	outCh      chan []byte
}

var (
	chanDM consChannel
	chanQA consChannel
)

func initConsChannels() {
	/* Init DM */
	chanDM.chConsName = opts.StrDataManager
	chanDM.inCh = make(chan []byte)
	chanDM.outCh = make(chan []byte)

	/* Init QA */
	chanQA.chConsName = opts.StrQueryAPI
	chanQA.inCh = make(chan []byte)
	chanQA.outCh = make(chan []byte)
}

//KafkaConsumer constructs Kafka-Consumer based on confluent-kafka-go library
func KafkaConsumer() error {
	brokers := opts.KafkaBrokerList
	topic := opts.KafkaTopic

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	opts.Logger.Println("Starting Kafka Consumer for topic", opts.KafkaTopic)
	k, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":               brokers,
		"group.id":                        opts.StrKafkaConGroupID,
		"go.events.channel.enable":        true,
		"go.application.rebalance.enable": true,
		"default.topic.config": kafka.ConfigMap{
			"auto.offset.reset": "earliest",
		},
	})
	if err != nil {
		opts.Logger.Fatalln("Failed to create kafka-consumer ", err)
	}
	k.Subscribe(topic, nil)
	manageChannels()
	registerMsgHandlers()

	doneCh := make(chan struct{})
	go func() {
		for {
			select {
			case sig := <-signalCh:
				opts.Logger.Println("Interrupt is detected", sig)
				doneCh <- struct{}{}
			case ev := <-k.Events():
				switch e := ev.(type) {
				case kafka.AssignedPartitions:
					opts.Logger.Println(e)
					k.Assign(e.Partitions)
				case kafka.RevokedPartitions:
					opts.Logger.Println(e)
					k.Unassign()
				case *kafka.Message:
					if opts.Verbose {
						opts.Logger.Printf("Received on [%s] messages %s\n", e.TopicPartition, string(e.Value))
					}
					if opts.Verbose {
						opts.Logger.Println("Txing inCh")
					}
					sendToInChannels(e.Value)
				case kafka.Error:
					opts.Logger.Println(e)
				}
			}
		}
	}()

	<-doneCh
	opts.Logger.Println("Kafka-Consumer Closed")
	return nil
}

func manageChannels() {
	initConsChannels()
	manageOutChannels(chanDM.inCh, chanDM.outCh)
	manageOutChannels(chanQA.inCh, chanQA.outCh)
}

func sendToInChannels(msg []byte) {
	if opts.SendToDM {
		chanDM.inCh <- msg
	}
	if opts.SendToQA {
		chanQA.inCh <- msg
	}
}

func manageOutChannels(inCh chan []byte, outCh chan []byte) {
	go func() {
		var inQueue [][]byte
		outChannel := func() chan []byte {
			if len(inQueue) == 0 {
				return nil
			}
			return outCh
		}
		enQueue := func(val []byte) {
			inQueue = append(inQueue, val)
			if opts.Verbose {
				opts.Logger.Println("Data Queued:", string(val))
			}
		}
		deQueue := func() []byte {
			if len(inQueue) == 0 {
				return nil
			}
			curVal := inQueue[0]
			inQueue = inQueue[1:]
			if opts.Verbose {
				opts.Logger.Println("Data served from Queue:", string(curVal))
			}
			return curVal
		}
		for len(inQueue) > 0 || inCh != nil {
			select {
			case v, ok := <-inCh:
				if opts.Verbose {
					opts.Logger.Println("Rxing in inCh")
				}
				if !ok {
					inCh = nil
				} else {
					enQueue(v)
				}
			case outChannel() <- deQueue():
			}
		}
		//opts.Logger.Println("Closing outCh as:", outCh)
		//close(outCh)
	}()
}

func registerMsgHandlers() {
	// Register all the message handlers here
	conChannelList := []consChannel{chanDM, chanQA}
	conChannelLen := len(conChannelList)
	for i := 0; i < conChannelLen; i++ {
		go func(i int) {
			mh := msghandler.NewMsgHandler(conChannelList[i].chConsName)
			mh.MHChan = conChannelList[i].outCh
			if err := mh.Run(); err != nil {
				opts.Logger.Fatalf("msgHandler run error %v ", err)
			}
		}(i)
	}
}
