package grpcclient

import (
	"time"
	"zinx/utils"
	"zinx/zlog"

	"go.etcd.io/etcd/clientv3"
)

type GrpcClient struct {
	//Grpc客户端句柄
	GrpcCnn *clientv3.Client
	GrpcKv  clientv3.KV
}

/*
	定义一个全局的对象
*/
var GrpcClientObject *GrpcClient

/*
	提供init方法，默认加载
*/
func init() {
	// 建立连接
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   utils.GlobalObject.EtcdIps,
		DialTimeout: utils.GlobalObject.EtcdDialTimeout * time.Second,
	})
	if err != nil {
		zlog.Error("etcd server Conn err = ", err)
		return
	}
	kvc := clientv3.NewKV(cli)
	//初始化GlobalObject变量，设置一些默认值
	GrpcClientObject = &GrpcClient{
		GrpcCnn: cli,
		GrpcKv:  kvc,
	}
}
