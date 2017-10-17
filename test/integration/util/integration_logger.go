// +build integration

/*
Copyright (C) 2017 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"time"
)

const (
	messageInfoMaxLength = 7
	timeHeaderLength     = 16
)

var (
	IntegrationLog *log.Logger
	logFile        *os.File
)

func init() {
	// Make sure there is a log, even before StartLog is called
	IntegrationLog = log.New(ioutil.Discard, "", 0)
}

func StartLog(logPath string) error {
	t := time.Now()
	logFileName := fmt.Sprintf("integration_%d-%d-%d_%02d-%02d-%02d.log", t.Year(), t.Month(),
		t.Day(), t.Hour(), t.Minute(), t.Second())
	logPath = path.Join(logPath, logFileName)
	logFile, err := os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	IntegrationLog = log.New(logFile, "", log.Ltime|log.Lmicroseconds)
	LogMessage("info", "Log Initiated")
	fmt.Println("Log successfully started, logging into:", logPath)

	return nil
}

func CloseLog() error {
	return logFile.Close()
}

func LogMessage(messageInfo, message string) error {
	messageInfo = formatMessageInfo(messageInfo)
	message = formatMessage(message)
	IntegrationLog.Print(messageInfo + message)

	return nil
}

func formatMessage(message string) string {
	formattedMessage := ": "
	offsetLength := timeHeaderLength + messageInfoMaxLength + len(formattedMessage)

	splittedMessage := strings.Split(message, "\n")
	for index, line := range splittedMessage {
		switch index {
		case 0:
			formattedMessage += line + "\n"
		case (len(splittedMessage) - 1):
			if line == "" {
				continue
			}
			fallthrough
		default:
			offSet := strings.Repeat(" ", offsetLength)
			formattedMessage += offSet + line + "\n"
		}
	}

	return formattedMessage
}

func formatMessageInfo(messageInfo string) string {
	if len(messageInfo) > messageInfoMaxLength {
		messageInfo = messageInfo[0:messageInfoMaxLength]
	} else if len(messageInfo) < messageInfoMaxLength {
		difference := messageInfoMaxLength - len(messageInfo)
		messageInfo = strings.Repeat(" ", difference) + messageInfo
	}

	return messageInfo
}
