package svc

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
)

type action func([]string) error

type Options struct {
	ptrHookAction  action
	aaaaHookAction action
	aHookAction    action
	whitelistMap   wbmap
	blacklistMap   wbmap
}

type Option func(opts *Options)

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

func WithSaveBList(blist string) Option {
	return func(opts *Options) {
		opts.blacklistMap = make(wbmap)
		opts.blacklistMap.saveCache(blist)
		fmt.Fprint(os.Stdout, opts.blacklistMap)
	}
}

func WithSaveWList(wlist string) Option {
	return func(opts *Options) {
		opts.whitelistMap = make(wbmap)
		opts.whitelistMap.saveCache(wlist)
		fmt.Fprint(os.Stdout, opts.whitelistMap)
	}
}
func WithSaveBWList(blist, wlist string) Option {
	return func(opts *Options) {
		opts.blacklistMap = make(wbmap)
		opts.whitelistMap = make(wbmap)
		opts.blacklistMap.saveCache(blist)
		opts.whitelistMap.saveCache(wlist)
		fmt.Fprint(os.Stdout, opts.blacklistMap, opts.whitelistMap)
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
