package model

type Config struct {
	Server     ServerConfig     `ini:"server"`
	Mysql      MysqlConfig      `ini:"mysql"`
	Logger     LoggerConfig     `ini:"logger"`
	Redis      RedisConfig      `ini:"redis"`
	Kafka      KafkaConfig      `ini:"kafka"`
	Etcd       EtcdConfig       `ini:"etcd"`
	Encrypter  EncrypterConfig  `ini:"encrypter"`
	Threadpool ThreadpoolConfig `ini:"threadpool"`
	RateLimite RateLimiteConfig `ini:"ratelimit"`
	UserGRPC   UserGRPCConfig   `ini:"usergrpc"`
	OrderGRPC  OrderGRPCConfig  `ini:"ordergrpc"`
	ItemGRPC   ItemGRPCConfig   `ini:"itemgrpc"`
}

type ServerConfig struct {
	Ip        string `ini:"ip"`
	Port      string `ini:"port"`
	Name      string `ini:"name"`
	Version   string `ini:"version"`
	Configkey string `ini:"configkey"`
	Machineid string `ini:"machineid"`
}

type MysqlConfig struct {
	Username string  `ini:"username"`
	Passwd   string  `ini:"passwd"`
	Database string  `ini:"database"`
	Host     string  `ini:"host"`
	Port     int     `ini:"port"`
	Timeout  float32 `ini:"timeout"`
}

type LoggerConfig struct {
	Type      string `ini:"type"`
	Path      string `ini:"path"`
	Name      string `ini:"name"`
	Level     string `ini:"level"`
	SplitType string `ini:"split_type"`
	ChanSize  string `ini:"chan_size"`
	SplitSize string `ini:"split_size"`
}

type RedisConfig struct {
	Provider    string `ini:"provider"`
	Ip          string `ini:"ip"`
	Password    string `ini:"password"`
	MaxIdle     int    `ini:"maxIdle"`
	MaxActive   int    `ini:"maxActive"`
	IdleTimeout int    `ini:"idleTimeout"`
	LockName    string `ini:"lockname"`
}

type KafkaConfig struct {
	Port  string `ini:"port"`
	Topic string `ini:"topic"`
}

type EtcdConfig struct {
	Ip          string `ini:"ip"`
	DialTimeout string `ini:"dialTimeout"`
}

type EncrypterConfig struct {
	Tool string `ini:"tool"`
}

type ThreadpoolConfig struct {
	Size string `ini:"size"`
}

type RateLimiteConfig struct {
	Threhold int64  `ini:"threhold"`
	Period   int    `ini:"period"`
	LockName string `ini:"lockname"`
	LockKey  string `ini:"lockkey"`
}

type UserGRPCConfig struct {
	ServerName string `ini:"servername"`
	ClientName string `ini:"clientname"`
}

type OrderGRPCConfig struct {
	ServerName string `ini:"servername"`
	ClientName string `ini:"clientname"`
}

type ItemGRPCConfig struct {
	ServerName string `ini:"servername"`
	ClientName string `ini:"clientname"`
}
