package cache

import (
	"fmt"
	"karst/utils"
	"sync"
	"time"
)

var lockCache uint64 = 0
var lock sync.RWMutex = sync.RWMutex{}
var basePath string = ""

func SetBasePath(tmpBasePath string) {
	lock.Lock()
	defer lock.Unlock()
	basePath = tmpBasePath
	lockCache = 0
}

func WaitLock(size uint64) error {
	for i := 0; i < 1000; i++ {
		canLock, err := Lock(size)
		if err != nil {
			return err
		}

		if canLock {
			return nil
		}

		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("Get cache timeout")
}

func Lock(size uint64) (bool, error) {
	lock.Lock()
	defer lock.Unlock()
	diskUsage, _ := utils.NewDiskUsage(basePath)
	if size >= diskUsage.Free+lockCache {
		return false, fmt.Errorf("Locked space is too large, need: %d, free: %d, lock: %d", size, diskUsage.Free, lockCache)
	}

	if size < diskUsage.Free {
		lockCache = lockCache + size
		return true, nil
	}

	return false, nil
}

func Unlock(size uint64) {
	lock.Lock()
	defer lock.Unlock()

	if lockCache < size {
		lockCache = 0
	} else {
		lockCache = lockCache - size
	}
}
