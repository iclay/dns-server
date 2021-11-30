## 运行：
   nohup ./server serve -c ../conf/confile 2<&1 &
## 使用：
   修改dns-client的dns服务器地址为dns-server的ip即可
## 部署目录结构描述：
```
.
├── bin
│   ├── nohup.out
│   └── server //二进制文件
├── conf
│   ├── confile //基础配置文件
│   └── login //api调用配置文件
├── log
│   ├── blog //黑名单日志记录文件
│   ├── log //日志文件
│   └── wlog //白名单日志记录文件
├── store
│   ├── store //dns缓存记录
│   └── store_bk //dns缓存记录备份
├
└── white
    └── whitelist //自定义白名单,支持正则表达
```

## 功能:
- [x] DNS server
  - [x] DNS forwarding
  - [x] DNS caching
  - [x] A record
  - [x] PTR record
  - [x] AAAA record
  - [x] CNAME record
  - [ ] NS record
  - [ ] SOA record
  - [ ] MX record
  - [ ] SRV record
- [x] REST server
  - [x] Create records
  - [x] Read records
  - [x] Update records
  - [x] Delete records
- [x] Filter
  - [x] Whitelist filter


   