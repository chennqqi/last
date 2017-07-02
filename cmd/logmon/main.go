package main

import (
	"fmt"
	"os/user"
	"time"

	"github.com/chennqqi/last"
)

func main() {
	ul, err := last.NewUnixLogMonitor("")
	if err != nil {
		fmt.Println("NewUnix log monitor error:", err)
		return
	}
	defer ul.Cleanup()
	for {
		log, ok := <-ul.LastLog
		if !ok {
			fmt.Println("quit")
			break
		}
		user, err := user.LookupId(fmt.Sprintf("%d", log.Uid))
		if err != nil {
			fmt.Println("uid", log.Uid, "name not exist", err)
			fmt.Printf("NEWLOG: %s %s @%s\n", log.Host, log.Line, time.Unix(log.Time, 0))
		} else {
			fmt.Printf("NEWLOG: %s#%s %s %s @%s\n", user.Name, user.Username, log.Host, log.Line, time.Unix(log.Time, 0))
		}
	}
}
