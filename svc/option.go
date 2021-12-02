package svc

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
)

type action func([]string) error

type Options struct {
	ptrHookAction  action
	aaaaHookAction action
	aHookAction    action
	hookAction     action
	whitelistMap   wbmap
	blacklistMap   wbmap
}

type Option func(opts *Options)

func registerSingal(fns ...func()) {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGUSR1)
	go func() {
		for {
			select {
			case <-c:
				fmt.Fprint(os.Stdout, "now reload whitelist or blacklist\n")
				for _, fn := range fns {
					fn()
				}
			}
		}
	}()
}

func WithPTRHookAction(fn action) Option {
	return func(opts *Options) {
		opts.ptrHookAction = fn
	}
}

func WithAAAAHookAction(fn action) Option {
	return func(opts *Options) {
		opts.aaaaHookAction = fn
	}
}

func WithAHookAction(fn action) Option {
	return func(opts *Options) {
		opts.aHookAction = fn
	}
}

func WithHookAction(fn action) Option {
	return func(opts *Options) {
		opts.hookAction = fn
	}
}

func WithSaveBList(blist string) Option {
	return func(opts *Options) {
		opts.blacklistMap = make(wbmap)
		opts.blacklistMap.saveCache(blist)
		go registerSingal(func() {
			opts.blacklistMap.clear()
			opts.blacklistMap.saveCache(blist)
			fmt.Fprint(os.Stdout, fmt.Sprintf("reload blacklist_map=%+v\n", opts.blacklistMap))
		})
		fmt.Fprint(os.Stdout, fmt.Sprintf("blacklist_map=%+v\n", opts.blacklistMap))
	}
}

func WithSaveWList(wlist string) Option {
	return func(opts *Options) {
		opts.whitelistMap = make(wbmap)
		opts.whitelistMap.saveCache(wlist)
		go registerSingal(func() {
			opts.whitelistMap.clear()
			opts.whitelistMap.saveCache(wlist)
			fmt.Fprint(os.Stdout, fmt.Sprintf("reload whitelist_map=%+v\n", opts.whitelistMap))

		})
		fmt.Fprint(os.Stdout, fmt.Sprintf("white_listmap=%+v\n", opts.whitelistMap))

	}
}
func WithSaveBWList(blist, wlist string) Option {
	return func(opts *Options) {
		opts.blacklistMap = make(wbmap)
		opts.whitelistMap = make(wbmap)
		opts.blacklistMap.saveCache(blist)
		opts.whitelistMap.saveCache(wlist)
		go registerSingal(func() {
			opts.blacklistMap.clear()
			opts.whitelistMap.clear()
			opts.blacklistMap.saveCache(blist)
			opts.whitelistMap.saveCache(wlist)
			fmt.Fprint(os.Stdout, fmt.Sprintf("reload whitelist_map=%+v, blacklist_map=%+v\n", opts.whitelistMap, opts.blacklistMap))
		})
		fmt.Fprint(os.Stdout, fmt.Sprintf("whitelist_map=%+v, blacklist_map=%+v\n", opts.whitelistMap, opts.blacklistMap))
	}
}

func loadOptions(options ...Option) *Options {
	opts := new(Options)
	for _, option := range options {
		option(opts)
	}
	return opts
}

type wbmap map[string]interface{}

func (wb wbmap) queryFiles(folder string) (files []string) {
	var fn func(string)
	fn = func(fold string) {
		elements, _ := ioutil.ReadDir(fold)
		for _, elem := range elements {
			if elem.IsDir() {
				fn(fold + "/" + elem.Name())
			} else {
				files = append(files, fold+"/"+elem.Name())
			}
		}
	}
	fn(folder)
	return
}
func (wb wbmap) saveCache(dir string) {
	if dir == "" {
		return
	}
	files := wb.queryFiles(dir)
	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			panic(err)
		}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			domain := scanner.Text()
			if domain == "" || domain[0] == '#' || domain[0] == '[' {
				continue
			}
			if len(domain) > 0 {
				wb[domain] = struct{}{}
			}
		}
	}
}

func (wb wbmap) clear() {
	for k, _ := range wb {
		delete(wb, k)
	}
}
