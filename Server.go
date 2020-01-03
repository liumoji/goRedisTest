package main

import (
	"context"
	"time"
	"zinx/grpcclient"
	"zinx/ziface"
	"zinx/zlog"
	"zinx/znet"
)

//test 自定义路由
type TestRouter struct {
	znet.BaseRouter
}

//Ping Handle
func (this *TestRouter) Handle(request ziface.IRequest) {

	zlog.Debug("Call PingRouter Handle")
	//先读取客户端的数据，再回写ping...ping...ping
	var msg string
	tmp := request.GetData()
	for i := 0; i < len(tmp); i++ {
		msg += tmp[i]
	}
	zlog.Debug("recv from client : msgId=", request.GetMsgID(), ", data=", msg)
	//取值，设置超时为1秒
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	resp, err := grpcclient.GrpcClientObject.GrpcKv.Put(context.TODO(), "/etcd/con", "hello world")
	//操作完毕，取消etcd
	cancel()
	if err != nil {
		zlog.Error("put etcd failed, err:", err)
		return
	}

	sendMsg := []string{"moji server ok"}
	err = request.GetConnection().SendMsg(0, sendMsg)
	if err != nil {
		zlog.Error(err)
	}
}

// type HelloZinxRouter struct {
// 	znet.BaseRouter
// }

// //HelloZinxRouter Handle
// func (this *HelloZinxRouter) Handle(request ziface.IRequest) {
// 	zlog.Debug("Call HelloZinxRouter Handle")
// 	//先读取客户端的数据，再回写ping...ping...ping
// 	zlog.Debug("recv from client : msgId=", request.GetMsgID(), ", data=", string(request.GetData()))

// 	err := request.GetConnection().SendBuffMsg(1, []byte("Hello Zinx Router V0.10"))
// 	if err != nil {
// 		zlog.Error(err)
// 	}
// }

//创建连接的时候执行
func DoConnectionBegin(conn ziface.IConnection) {
	zlog.Debug("DoConnecionBegin is Called ... ")

	//设置两个链接属性，在连接创建之后
	zlog.Debug("Set conn Name, Home done!")
	// conn.SetProperty("Name", "Aceld")
	// conn.SetProperty("Home", "https://www.jianshu.com/u/35261429b7f1")

	sendMsg := []string{"moji server connect ok"}
	err := conn.SendMsg(2, sendMsg)
	if err != nil {
		zlog.Error(err)
	}
}

//连接断开的时候执行
func DoConnectionLost(conn ziface.IConnection) {
	//在连接销毁之前，查询conn的Name，Home属性
	if name, err := conn.GetProperty("Name"); err == nil {
		zlog.Error("Conn Property Name = ", name)
	}

	if home, err := conn.GetProperty("Home"); err == nil {
		zlog.Error("Conn Property Home = ", home)
	}

	zlog.Debug(conn.RemoteAddr().String(), "DoConneciotnLost is Called ... ")
}

func main() {
	defer grpcclient.GrpcClientObject.GrpcCnn.Close()
	//创建一个server句柄
	s := znet.NewServer()

	//注册链接hook回调函数
	s.SetOnConnStart(DoConnectionBegin)
	s.SetOnConnStop(DoConnectionLost)

	//配置路由
	s.AddRouter(0, &TestRouter{})
	//s.AddRouter(1, &HelloZinxRouter{})

	//开启服务
	s.Serve()
}
