package main

/*
  Copyright (c) IBM Corporation 2016

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific

   Contributors:
     Mark Taylor - Initial Contribution
*/

import (
	log "github.com/Sirupsen/logrus"
	"github.com/ibm-messaging/mq-golang/mqmetric"
	"os"
	"time"
)

func initLog() {
	level, err := log.ParseLevel(config.logLevel)
	if err != nil {
		level = log.InfoLevel
	}
	log.SetLevel(level)
	// Since this program prints its data to stdout, need any
	// log info to go elsewhere.
	log.SetOutput(os.Stderr)
}

func main() {
	var err error

	initConfig()

	initLog()

	if config.qMgrName == "" {
		log.Errorln("Must provide a queue manager name to connect to.")
		os.Exit(1)
	}
	d, err := time.ParseDuration(config.interval + "s")
	if err != nil {
		log.Errorln("Invalid value for interval parameter: ", err)
		os.Exit(1)
	}

	log.Infoln("Starting IBM MQ metrics exporter for JSON")

	// Connect and open standard queues
	err = mqmetric.InitConnection(config.qMgrName, config.replyQ, &config.cc)
	if err == nil {
		log.Infoln("Connected to queue manager ", config.qMgrName)
		defer mqmetric.EndConnection()
	}

	// What metrics can the queue manager provide? Find out, and
	// subscribe.
	if err == nil {
		err = mqmetric.DiscoverAndSubscribe(config.monitoredQueues, true, "")
	}

	// Go into main loop for sending data to stdout
	// This program runs forever
	if err == nil {
		for {
			time.Sleep(d)
			Collect(getNRConfig())
			// time.Sleep(d)
			// fmt.Println("After")
		}
	}

	if err != nil {
		log.Fatal(err)
	}

	os.Exit(0)
}
