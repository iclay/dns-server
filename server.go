package main

import (
	"dns/api"
	"dns/custom"
	"dns/logwriter"
	"dns/svc"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
)

var Version = "manual build has no version"

func main() {

	app := &cli.App{
		EnableBashCompletion: true,
		Name:                 "DNS",
		Usage:                "Domain name resolution",
		Commands: []*cli.Command{
			{
				Name:   "serve",
				Usage:  "start the server",
				Action: serve,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "config",
						Usage:    "config file",
						Required: true,
						Aliases:  []string{"c"},
						EnvVars:  []string{"DNS_SERVER_CONFIG"},
					},
				},
			},
		},
		Authors: []*cli.Author{
			{
				Name:  "Tsinglink tech",
				Email: "tech@qinglianyun.com",
			},
		},
		Copyright: "Beijing Tsinglink Cloud Technology Co., Ltd (2021)",
		Version:   Version,
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func serve(c *cli.Context) error {
	flag.Parse()
	svc.ParseFile(c.String("config"))
	svc.Print()
	GConf := svc.GCONF
	logMap := make(map[string]*logrus.Logger)
	for _, v := range []string{"log", "wlog" /*白名单日志*/, "blog" /*黑名单日志*/} {
		lw := &logwriter.HourlySplit{
			Dir:           GConf.LogPath,
			FileFormat:    v + "_2006-01-02T15",
			MaxFileNumber: int64(GConf.LogMaxFileNum),
			MaxDiskUsage:  GConf.LogMaxDiskUsage,
		}
		defer lw.Close()

		lg := log.New()
		customFormatter := &log.TextFormatter{
			TimestampFormat: time.RFC3339,
			FullTimestamp:   true,
		}

		lg.SetFormatter(customFormatter)
		lg.SetReportCaller(true)
		lg.SetOutput(lw)
		lv, err := log.ParseLevel(GConf.LogLevel)
		if err != nil {
			lv = log.WarnLevel
		}
		lg.SetLevel(lv)
		logMap[v] = lg
	}
	for _, v := range []string{"log", "wlog", "blog"} {
		if logMap[v] == nil {
			panic("init log error")
		}
	}
	lg := logMap["log"]
	if err := os.MkdirAll(GConf.RWDirPath, 0666); err != nil {
		lg.Errorf("create rwdirpath: %v error: %v", GConf, err)
		return err
	}
	lg.Info("start dns server")
	svc.SetLogger(logMap)
	api.SetClient(&api.Client{
		AuthLogin: &api.Login{
			AccountType: GConf.AccountType,
			GrantType:   GConf.GrantType,
			Email:       GConf.Email,
			Password:    GConf.Password,
		},
		RemoteHost: GConf.RemoteHost,
		MaxTry:     3,
	})
	dns := svc.NewDNService(GConf.RWDirPath, []net.UDPAddr{{IP: net.ParseIP(GConf.ForwardIP), Port: GConf.ForwardPort}},
		svc.WithPTRHookAction(custom.PTRHookAction),
		svc.WithAAAAHookAction(custom.AAAAHookAction),
		svc.WithAHookAction(custom.AHookAction),
		svc.WithSaveWList(GConf.WhiteList),
	)
	rest := svc.RestService{Dn: dns}
	//通过restfulapi的调用支持添加，读取，更新，删除功能
	dnsHandler := func() http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodPost:
				rest.Create(w, r)
			case http.MethodGet:
				rest.Read(w, r)
			case http.MethodPut:
				rest.Update(w, r)
			case http.MethodDelete:
				rest.Delete(w, r)
			}
		}
	}

	withAuth := func(h http.HandlerFunc) http.HandlerFunc {
		var _ = "intercept"
		return func(w http.ResponseWriter, r *http.Request) {
			h(w, r)
		}
	}

	http.Handle("/dns", withAuth(dnsHandler()))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", GConf.ServerPort), nil))
	return nil
}
