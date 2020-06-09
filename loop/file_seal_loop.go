package loop

import (
	"fmt"
	"karst/fs"
	"karst/model"
	"karst/tee"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
)

const (
	fileSealJobQueueLimit = 1000
)

var fileSealJobs chan model.FileSealMessage = nil

func StartFileSealLoop(db *leveldb.DB, fs fs.FsInterface, tee *tee.Tee) {
	// Seal jobs queue
	fileSealJobs = make(chan model.FileSealMessage, fileSealJobQueueLimit)
	go fileSealLoop(db, fs, tee)
}

func TryEnqueueFileSealJob(job model.FileSealMessage) bool {
	if fileSealJobs == nil {
		return false
	}

	select {
	case fileSealJobs <- job:
		return true
	default:
		return false
	}
}

func fileSealLoop(db *leveldb.DB, fs fs.FsInterface, tee *tee.Tee) {
	for {
		select {
		case job := <-fileSealJobs:
			fmt.Printf("File seal job: %s \n", job.StoreOrderHash)
			time.Sleep(1 * time.Second)
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}
