## 运行：
   bash start.sh start //启动服务  
   bash start.sh stop //关闭服务  
   bash start.sh restart //重启服务  
   bash start.sh update //服务自动重新加载白名单(如果修改了白名单，只需执行此命令即可)
## 使用：
   修改dns-client的dns服务器地址为dns-server的ip即可
## 部署目录结构描述：
```
├── bin
│   ├── nohup.out
│   ├── server //二进制文件
│   └── start.sh //启动脚本,使用(bash start.sh -h)命令，查看详细使用说明
├── conf
│   ├── confile //基础配置文件
│   └── login //api调用配置文件
├── log
│   ├── blog //黑名单日志记录文件
│   ├── log //日志文件
│   └── wlog ///白名单日志记录文件
├── README.md
├── store
│   ├── store //dns缓存记录
│   └── store_bk //dns缓存记录备份
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


## DNS记录变更:
```shell
// 新增A记录    
curl -X POST http://localhost:10001/dns -H 'Content-Type: application/json' -d '{"Host":"example_test.com.","TTL": 600,"Type":"A","Data":"192.168.1.1"}'  
// 修改A记录  
curl -X PUT http://localhost:10001/dns -H 'Content-Type: application/json' -d ' {"Host":"example_test.com.","TTL": 600,"Type": "A","OldData":"192.168.1.1","Data":"192.168.1.2"}'  
// 删除A记录  
curl -X DELETE http://localhost:10001/dns -H 'Content-Type: application/json' -d '{"Host":"example_test.com.","Type": "A"}'  
```
