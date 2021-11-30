package sock

import (
	"os"
	"runtime"

	"golang.org/x/sys/unix"
)

type epollReactor struct {
	ln          *listener
	main_poller poller
	sub_poller  []poller
	opt         Option
}

type Option struct {
	reuseport    bool //是否开启reuseport
	numEventLoop int  //核数
	multiCore    bool //是否开启多核
}

type poller struct {
	fd    int
	index int
	ln    *listener
}

const (
	readEvents      = unix.EPOLLPRI | unix.EPOLLIN
	writeEvents     = unix.EPOLLOUT
	readWriteEvents = readEvents | writeEvents
)

func (p *poller) AddRead(pa *PollAttachment) error {
	return os.NewSyscallError("epoll_ctl_add", unix.EpollCtl(p.fd, unix.EPOLL_CTL_ADD, pa.FD, &unix.EpollEvent{Fd: int32(pa.FD), Events: readEvents}))
}

func (e *epollReactor) Service() {
	numEventLoop := 1
	if e.opt.multiCore {
		numEventLoop = runtime.NumCPU()
	}
	e.start(numEventLoop)
}

type PollAttachment struct {
	FD       int
	Callback PollEventHandler
}
type PollEventHandler func(uint32) error

func (e *epollReactor) start(num int) (err error) {
	for i := 0; i < num; i++ {
		sub_poller := new(poller)
		if sub_poller.fd, err = unix.EpollCreate1(unix.EPOLL_CLOEXEC); err != nil {
			err = os.NewSyscallError("sub_poller epoll_create1 error:", err)
			return
		}
		sub_poller.index = i
		e.sub_poller = append(e.sub_poller, *sub_poller)
	}
	main_poller := new(poller)
	if main_poller.fd, err = unix.EpollCreate1(unix.EPOLL_CLOEXEC); err != nil {
		err = os.NewSyscallError("main_poller epoll_create1 error:", err)
		return
	}
	main_poller.index = -1
	main_poller.ln = e.ln
	e.main_poller = *main_poller
	if err = e.main_poller.AddRead(&PollAttachment{FD: e.ln.fd}); err != nil {
		return
	}
	return err
}

// func (rpoll_reacti)
// poll(fd, ca) //给fd添加一个函数

// epoll_polling(epoll_event, fd, func(fd)) error {

// }
// acctptor() {
// 	if fd == listenfd {
// 		添加到epoll的等待队列
// 	}
// }

// 进程1 进程2 进程3   53号端口

// 1           3

//epoll1_in epoll2_in epoll3_in
