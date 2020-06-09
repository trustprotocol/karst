package loop

import (
	"fmt"
	"karst/model"
	"time"
)

const (
	fileSealJobQueueLimit = 1000
)

var sealJobs chan model.FileSealMessage = nil

func StartFileSealLoop() {
	// Seal jobs queue
	sealJobs = make(chan model.FileSealMessage, fileSealJobQueueLimit)
	go fileSealLoop(sealJobs)
}

func TryEnqueueFileSealJob(job model.FileSealMessage) bool {
	if sealJobs == nil {
		return false
	}

	select {
	case sealJobs <- job:
		return true
	default:
		return false
	}
}

func fileSealLoop(sealJobs <-chan model.FileSealMessage) {
	for {
		select {
		case job := <-sealJobs:
			fmt.Printf("File seal job: %s \n", job.StoreOrderHash)
			time.Sleep(1 * time.Second)
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}
