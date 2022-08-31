package bucket

import (
	"sync"

	"golang.org/x/sync/errgroup"
)

type Storage struct {
	loginMaxLoad, passwordMaxLoad, ipMaxLoad int
	loginMx, passwordMx, ipMx                sync.Mutex
	loginBuckets, passwordBuckets, ipBuckets map[string]*Bucket
	loginDelCh, passwordDelCh, ipDelCh       chan string
}

type Config interface {
	GetLoginMaxLoad() int
	GetPasswordMaxLoad() int
	GetIPMaxLoad() int
}

func NewStorage(config Config) *Storage {
	s := &Storage{
		loginMaxLoad:    config.GetLoginMaxLoad(),
		passwordMaxLoad: config.GetPasswordMaxLoad(),
		ipMaxLoad:       config.GetIPMaxLoad(),

		loginBuckets:    map[string]*Bucket{},
		passwordBuckets: map[string]*Bucket{},
		ipBuckets:       map[string]*Bucket{},

		loginDelCh:    make(chan string),
		passwordDelCh: make(chan string),
		ipDelCh:       make(chan string),
	}

	go s.storageCleaner()

	return s
}

func (s *Storage) Clear(login, password, ip string) {
	s.loginMx.Lock()
	delete(s.loginBuckets, login)
	s.loginMx.Unlock()

	s.passwordMx.Lock()
	delete(s.passwordBuckets, password)
	s.passwordMx.Unlock()

	s.ipMx.Lock()
	delete(s.ipBuckets, ip)
	s.ipMx.Unlock()
}

func (s *Storage) storageCleaner() {
	for {
		select {
		case login := <-s.loginDelCh:
			s.loginMx.Lock()
			delete(s.loginBuckets, login)
			s.loginMx.Unlock()
		case password := <-s.passwordDelCh:
			s.passwordMx.Lock()
			delete(s.passwordBuckets, password)
			s.passwordMx.Unlock()
		case ip := <-s.ipDelCh:
			s.ipMx.Lock()
			delete(s.ipBuckets, ip)
			s.ipMx.Unlock()
		}
	}
}

func (s *Storage) Check(login, password, ip string) error {
	return s.collectErrors(login, password, ip)
}

func (s *Storage) collectErrors(login, password, ip string) error {
	var eg errgroup.Group

	eg.Go(func() error {
		s.loginMx.Lock()
		defer s.loginMx.Unlock()

		return checkBucket(s.loginBuckets, login, s.loginMaxLoad, s.loginDelCh)
	})

	eg.Go(func() error {
		s.passwordMx.Lock()
		defer s.passwordMx.Unlock()

		return checkBucket(s.passwordBuckets, password, s.passwordMaxLoad, s.passwordDelCh)
	})

	eg.Go(func() error {
		s.ipMx.Lock()
		defer s.ipMx.Unlock()

		return checkBucket(s.ipBuckets, ip, s.ipMaxLoad, s.ipDelCh)
	})

	return eg.Wait()
}

func checkBucket(buckets map[string]*Bucket, key string, maxLoad int, delCh chan<- string) error {
	b, ok := buckets[key]
	if !ok {
		b = NewBucket(key, maxLoad, delCh)
		buckets[key] = b
	}

	return b.check()
}
