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

/*
The Collect() function is the key operation
invoked at the configured intervals, causing us to read available publications
and update the various data points.
*/

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/ibm-messaging/mq-golang/mqmetric"
	"go-insights-v2/client"
	"os"
	"strings"
)

var (
	first     = true
	eventData map[string]interface{}
)

/*
Collect is called by the main routine at regular intervals to provide current
data
*/
func Collect(insertKey string, accountNumber string) error {
	var err error
	eventData = make(map[string]interface{})

	log.Infof("IBM MQ JSON collector started")

	// Clear out everything we know so far. In particular, replace
	// the map of values for each object so the collection starts
	// clean.
	for _, cl := range mqmetric.Metrics.Classes {
		for _, ty := range cl.Types {
			for _, elem := range ty.Elements {
				elem.Values = make(map[string]int64)
			}
		}
	}

	// Process all the publications that have arrived
	mqmetric.ProcessPublications()

	// Have now processed all of the publications, and all the MQ-owned
	// value fields and maps have been updated.
	//
	// Now need to set all of the real items with the correct values
	//if first {
	// Always ignore the first loop through as there might
	// be accumulated stuff from a while ago, and lead to
	// a misleading range on graphs.
	//	first = false
	//} else {
	eventData["eventType"] = "MQSample"
	eventData["hostname"], _ = os.Hostname()
	eventData["queueManager"] = config.qMgrName

	for _, cl := range mqmetric.Metrics.Classes {
		for _, ty := range cl.Types {
			for _, elem := range ty.Elements {
				for key, value := range elem.Values {
					f := mqmetric.Normalise(elem, key, value)
					tags := map[string]string{
						"qmgr": config.qMgrName,
					}

					if key != mqmetric.QMgrMapKey {
						tags["object"] = key
					}
					printPoint(elem.MetricName, float32(f), tags)

				}
			}
		}
	}

	insights := client.NewInsertClient(insertKey, accountNumber)
	err = insights.PostEvent(eventData)

	if err != nil {
		fmt.Printf(" Error Posting %s/n", err)
	}
	//}

	return err

}

func printPoint(metric string, val float32, tags map[string]string) {

	if q, ok := tags["object"]; ok {
		eventData["queue"] = q
	}

	eventData[fixup(metric)] = val

	return
}

func fixup(s1 string) string {
	// Another reformatting of the metric name - this one converts
	// something like queue_avoided_bytes into queueAvoidedBytes
	s2 := ""
	c := ""
	nextCaseUpper := false

	for i := 0; i < len(s1); i++ {
		if s1[i] != '_' {
			if nextCaseUpper {
				c = strings.ToUpper(s1[i : i+1])
				nextCaseUpper = false
			} else {
				c = strings.ToLower(s1[i : i+1])
			}
			s2 += c
		} else {
			nextCaseUpper = true
		}

	}
	return s2
}
