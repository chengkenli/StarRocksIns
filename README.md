StarRocks Inspection

这是一个支持StarRocks自动巡检的工具，支持日志巡检、机器巡检、配置巡检、服务巡检、内表巡检...，然后生成巡检结论与评分。

![](C:\Users\c0l0f9l\SDK\StarRocksIns\img\2026-01-20-11-25-56-image.png)

# Config-yaml

```yaml
metadb:
  host: 127.0.0.1
  port: 3306
  user: root
  password: **************1.
  base: chengken.starrocks_information_connections

server:
  port: 19111
  loadhtmlglob: /u1/dlopsnas/chengken/insp/*
  loadstatic: /u1/dlopsnas/chengken/insp/static
  privateuser: starrocks
  privatekey: /u/users/starrocks/.ssh/id_rsa
  auditlog: audit.starrocks_audit_log

subject:
  log: chengken.starrocks_information_inspection_log
  avgs: chengken.starrocks_information_inspection_avgs
  conf: chengken.starrocks_information_inspection_conf
  olap: chengken.starrocks_information_inspection_olap



log:
  path: '/u/users/svccndlopsns/chengken/log'
```

## Options：

![](C:\Users\c0l0f9l\SDK\StarRocksIns\img\2026-01-20-11-29-41-image.png)

### Service

服务端：启用服务端后，支持直接在浏览器查看巡检报告（html体现）

```service
./StarRocksIns -service
```

### Agent

巡检端：仅做巡检工作，并且生成相关的报告文件。（最后交给Service来展示）

```agent
./StarRocksIns -agent
```

### Config

```配置文件解读
metadb:
  host: MySQL地址
  port: MySQL端口
  user: MySQL账户
  password: MySQL密码
  base: MySQL-StarRocks连接配置表，这里需要填写连接StarRocks的账号密码IP

server:
  port: 服务端口
  loadhtmlglob: 巡检报告生成的地址
  loadstatic: 巡检报告生成时需要的静态文件
  privateuser: ssh登录到fe、be节点的用户，免密（日志、参数、配置文件巡检需要）
  privatekey: ssh私钥文件（日志、参数、配置文件巡检需要）
  auditlog: 审计日志表

subject:
  log: 日志巡检，在表中填写指标后，将会忽略检查该指标（非必填）
  avgs: 参数巡检，在表中填写指标后，将会忽略检查该指标（非必填）
  conf: 配置巡检，在表中填写指标后，将会忽略检查该指标（非必填）
  olap: 内表巡检，在表中填写指标后，将会忽略检查该指标（非必填）
```

# Start

启动巡检：

![](C:\Users\c0l0f9l\SDK\StarRocksIns\img\2026-01-20-13-20-56-image.png)

```
./StarRocksIns -s <集群名称> -agent
./StarRocksIns -s sr-scct   -agent
```

![](C:\Users\c0l0f9l\SDK\StarRocksIns\img\2026-01-20-13-31-13-image.png)

访问报告：

```
./StarRocksIns -service
```

![](C:\Users\c0l0f9l\SDK\StarRocksIns\img\2026-01-20-13-33-00-image.png)

```
然后在浏览器打开：http://127.0.0.1:19111/sr-scct
```

首页

![](C:\Users\c0l0f9l\SDK\StarRocksIns\img\2026-01-20-13-35-08-image.png)

日志巡检

![](C:\Users\c0l0f9l\SDK\StarRocksIns\img\2026-01-20-13-36-08-image.png)

节点巡检

![](C:\Users\c0l0f9l\SDK\StarRocksIns\img\2026-01-20-13-36-42-image.png)

参数巡检

![](C:\Users\c0l0f9l\SDK\StarRocksIns\img\2026-01-20-13-37-22-image.png)

配置巡检

![](C:\Users\c0l0f9l\SDK\StarRocksIns\img\2026-01-20-13-37-48-image.png)

内表巡检

![](C:\Users\c0l0f9l\SDK\StarRocksIns\img\2026-01-20-13-43-14-image.png)
