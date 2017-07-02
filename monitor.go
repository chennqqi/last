/*
support unix like system only!!!
*/
package last

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/chennqqi/goutils/utils"
	"github.com/chennqqi/tail/watch"
	"gopkg.in/tomb.v1"
)

const (
	UNIX_LASTLOG = "/var/log/lastlog"
	UNIX_UIDMAX  = 65535
)

var ErrStop = errors.New("Err On Stop")
var PollMode = true

type UnixLogEvent struct {
	Uid int
	LastLogGo
}

type UnixLogMonitor struct {
	name    string
	watcher watch.FileWatcher
	changes *watch.FileChanges

	LastLog    chan *UnixLogEvent
	lastLogs   [UNIX_UIDMAX]LastLogGo
	firstCheck bool

	tomb.Tomb // provides: Done, Kill, Dying
	//	lk sync.Mutex
}

func NewUnixLogMonitor(name string) (*UnixLogMonitor, error) {
	if name == "" {
		name = UNIX_LASTLOG
	}

	var ul UnixLogMonitor
	if PollMode {
		ul.watcher = watch.NewPollingFileWatcher(name)
	} else {
		ul.watcher = watch.NewInotifyFileWatcher(name)
	}
	//to avoid block
	ul.LastLog = make(chan *UnixLogEvent, UNIX_UIDMAX)
	ul.name = name
	ul.firstCheck = true
	go ul.startMonitor()
	return &ul, nil
}

func (ul *UnixLogMonitor) doCheck() {
	var llgs [UNIX_UIDMAX]LastLogGo

	for i := 0; i < UNIX_UIDMAX; i++ {
		llg, err := ByUID(i)
		if err == ErrUnspportPlat {
			return
		}
		if err == nil {
			llgs[i] = llg
		}
	}
	if ul.firstCheck {
		ul.firstCheck = false
		return
	}

	for i := 0; i < UNIX_UIDMAX; i++ {
		if ul.lastLogs[i] != llgs[i] {
			ul.LastLog <- &UnixLogEvent{i, llgs[i]}
			ul.lastLogs[i] = llgs[i]
		}
	}
}

func (ul *UnixLogMonitor) startMonitor() error {
	defer ul.Done()
	defer close(ul.LastLog)

	//init data
	for i := 0; i < UNIX_UIDMAX; i++ {
		llg, err := ByUID(i)
		if err == ErrUnspportPlat {
			return err
		}
		if err == nil {
			ul.lastLogs[i] = llg
		}
	}

	reopen := func() error {
		for {
			var err error
			exist, err := utils.PathExists2(ul.name)
			if !exist {
				if os.IsNotExist(err) {
					log.Printf("Waiting for %s to appear...", ul.name)
					if err := ul.watcher.BlockUntilExists(&ul.Tomb); err != nil {
						if err == tomb.ErrDying {
							return err
						}
						return fmt.Errorf("Failed to detect creation of %s: %s", ul.name, err)
					}
					continue
				}
				return fmt.Errorf("Unable to open file %s: %s", ul.name, err)
			}
			break
		}
		return nil
	}
	// Read line by line.
	for {
		ul.doCheck()

		if ul.changes == nil {
			st, err := os.Stat(ul.name)
			if os.IsNotExist(err) {
				fmt.Println(ul.name, "is not exist")
				return err
			}

			ul.changes, err = ul.watcher.ChangeEvents(&ul.Tomb, st.Size())
			if err != nil {
				return err
			}
		}

		select {
		case <-ul.changes.Deleted:
			ul.changes = nil
			fmt.Println("deleted")
			if true {
				// XXX: we must not log from a library.
				log.Printf("Re-opening moved/deleted file %s ...", ul.name)
				if err := reopen(); err != nil {
					return err
				}
				log.Printf("Successfully reopened %s", ul.name)
			} else {
				log.Printf("Stopping tail as file no longer exists: %s", ul.name)
				return ErrStop
			}

		case <-ul.changes.Modified:
			fmt.Println("modified")

		case <-ul.changes.Truncated:
			fmt.Println("truncated")

			// Always reopen truncated files (Follow is true)
			log.Printf("Re-opening truncated file %s ...", ul.name)
			if err := reopen(); err != nil {
				return err
			}
			log.Printf("Successfully reopened truncated %s", ul.name)
		}
	}
}

func (ul *UnixLogMonitor) Cleanup() {
	watch.Cleanup(ul.name)
}

func (ul *UnixLogMonitor) Stop() error {
	ul.Kill(nil)
	return ul.Wait()
}
