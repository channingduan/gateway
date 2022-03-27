package config

type Config struct {
	// 访问地址(etcd,consul,zookeeper)
	Addr string
	// 服务注册地址
	RegistryAddr string
	// 服务目录
	BasePath string
	// 失败模型
	FailMode int
	// 轮询模型
	SelectMode int
}
