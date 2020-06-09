package loop

import (
	"context"
	"fmt"
	"karst/model"
	"time"
)

const (
	fileSealJobQueueLimit = 100
)

var StopFileSealLoop context.CancelFunc
var ctx context.Context

func StartFileSealLoop() {
	// Seal jobs queue
	sealJobs := make(chan model.FileSealMessage, fileSealJobQueueLimit)
	// Used to pause loop
	ctx, StopFileSealLoop = context.WithCancel(context.Background())

	go fileSealLoop(sealJobs, ctx)
}

func TryEnqueueFileSealJob(job model.FileSealMessage, sealJobs chan model.FileSealMessage) bool {
	select {
	case sealJobs <- job:
		return true
	default:
		return false
	}
}

func fileSealLoop(sealJobs <-chan model.FileSealMessage, ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case job := <-sealJobs:
			fmt.Printf("File seal job: %s \n", job.StoreOrderHash)
			time.Sleep(1 * time.Second)
		}
	}
}
