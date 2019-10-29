package main

import (
	"fmt"
	"io/ioutil"
	"runtime"
	"strconv"
	"time"
)

type hostsFileEntry struct {
	hostname       *string
	ip             *string
	ipResovledFrom *string
}

type hostsFileEntries []*hostsFileEntry

func hostsFileLocation() *string {
	var hostsFile string

	if runtime.GOOS == "windows" {
		hostsFile = "${SystemRoot}/System32/drivers/etc/hosts"
	} else {
		hostsFile = "/etc/hosts"
	}

	return &hostsFile
}

func createHostsBackup(hostsFileLocation *string) {
	input, err := ioutil.ReadFile(*hostsFileLocation)
	if err != nil {
		fmt.Println(err)
		return
	}

	backupLocation := *hostsFileLocation + ".backup-" + strconv.FormatInt(time.Now().Unix(), 10)

	err = ioutil.WriteFile(backupLocation, input, 0644)
	if err != nil {
		fmt.Println("Error creating", backupLocation)
		fmt.Println(err)
		return
	}
}
