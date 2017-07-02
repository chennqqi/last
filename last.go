package last

/*
#if !defined(_WIN32) && !defined(_WIN64)
#include <fcntl.h>
#include <unistd.h>
#include <stdlib.h>
#include <utmp.h>
struct lastlog* new_lastlog()
{
	struct lastlog* pNew = malloc(sizeof(struct lastlog));
	return pNew;
}
int get_lastlog(int uid, struct lastlog* pl)
{
	off_t pos = (off_t)uid * sizeof(struct lastlog);
	int fd = open(_PATH_LASTLOG, O_RDONLY, 0);
	if (fd >= 0) {
		ssize_t rxLen = pread(fd, pl, sizeof(struct lastlog), pos);
		close(fd);
		return rxLen==sizeof(struct lastlog)?0:-1;
	}
	return -1;
}

int lastlog_gettime(const struct lastlog* pl)
{
	return pl->ll_time;
}

const char* lastlog_gethost(const struct lastlog* pl)
{
	return pl->ll_host;
}

const char* lastlog_getline(const struct lastlog* pl)
{
	return pl->ll_line;
}
#else
#include <stdlib.h>
struct lastlog* new_lastlog()
{
	struct lastlog* pNew = malloc(1);
	return pNew;
}

int get_lastlog(int uid, struct lastlog* pl)
{
	return -2;
}

int lastlog_gettime(const struct lastlog* pl)
{
	return 0;
}

const char* lastlog_gethost(const struct lastlog* pl)
{
	return "";
}

const char* lastlog_getline(const struct lastlog* pl)
{
	return "";
}

#endif

*/
import "C"

import (
	"errors"
	"unsafe"
)

type LastLogGo struct {
	Time int64
	Host string
	Line string
}

var ErrNever = errors.New("never logged in")
var ErrMemory = errors.New("alloc memory error")
var ErrUnspportPlat = errors.New("only support unix like systemp")

func lastlogCToGo(pl *C.struct_lastlog) LastLogGo {
	var llg LastLogGo
	llg.Time = int64(int32(C.lastlog_gettime(pl)))
	llg.Line = C.GoString(C.lastlog_getline(pl))
	llg.Host = C.GoString(C.lastlog_gethost(pl))
	return llg
}

// ByUID returns last system login of user by UID
func ByUID(uid int) (LastLogGo, error) {
	var llg LastLogGo

	pl := C.new_lastlog()
	if pl == nil {
		return llg, ErrMemory
	}
	defer C.free(unsafe.Pointer(pl))

	r := C.get_lastlog(C.int(uid), pl)
	switch int(r) {
	case 0:
		return lastlogCToGo(pl), nil
	case -2:
		return llg, ErrUnspportPlat
	default:
	case -1:
		return llg, ErrNever
	}
	return llg, nil
}
