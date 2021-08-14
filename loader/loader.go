package loader

import (
	"io"
)


type Loader interface {
	LoadFile(fn string)
	Load(f io.ReadCloser)
	Shutdown(err error)
	GetChan() chan interface{}
	Abort()
}

type BaseLoader struct {
	c chan interface{}
	quitCh chan bool
}

func NewBaseLoader() *BaseLoader {
	return &BaseLoader{
		c: make(chan interface{}, 10),
		quitCh: make(chan bool, 2),
	}
}

func (l *BaseLoader) GetChan() chan interface{} {
	return l.c
}

func (l *BaseLoader) GetQuitChan() chan bool {
	return l.quitCh
}

func (l *BaseLoader) Abort() {
	quitCh := l.quitCh
	l.quitCh = nil
	if quitCh != nil {
		quitCh <- true
		close(quitCh)
		l.drain()
		for {
			_, ok := <-quitCh
			if !ok {
				break
			}
		}
	}
}

func (l *BaseLoader) drain() {
	c := l.c
	l.c = nil
	if c != nil {
		for {
			_, ok := <-c
			if !ok {
				break
			}
		}
		close(c)
	}
}

func (l *BaseLoader) Shutdown(err error) {
	c := l.c
	if c != nil {
		if err != nil {
			c <- err
		}
		close(c)
	}
}
