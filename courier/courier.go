package courier

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/duythinht/chaika/config"
	"github.com/hashicorp/consul/api"
)

type Courier interface {
	Send(serviceName string, catalog string, level string, message string)
	Close()
	GetHost() string
	GetPort() int
}

type LogInfo struct {
	Host string
	Port int
	Type string
}

var couriers map[string]Courier
var expired map[string]int64
var kv *api.KV

func Setup() {
	cfg := config.GetConfig()
	expired = make(map[string]int64)
	couriers = make(map[string]Courier)
	fmt.Println("Initilize couriers")

	client, err := api.NewClient(&api.Config{
		Address: cfg.ConsulHost + ":" + strconv.FormatInt(cfg.ConsulPort, 10),
	})
	//api.DefaultConfig())
	CheckError(err)
	// Get a handle to the KV API
	kv = client.KV()
}

func Get(serviceName string) Courier {

	now := time.Now().Unix()

	if expiredTime, ok := expired[serviceName]; ok && expiredTime > now {
		return couriers[serviceName]
	}

	logCfg := GetLogOutput(serviceName)

	if serv, ok := couriers[serviceName]; ok {
		if serv.GetHost() != logCfg.Host && serv.GetPort() != logCfg.Port {
			serv.Close()
			couriers[serviceName] = CreateGelf(serviceName, logCfg.Host, logCfg.Port)
		}
	} else {
		couriers[serviceName] = CreateGelf(serviceName, logCfg.Host, logCfg.Port)
	}

	expired[serviceName] = now + 5
	return couriers[serviceName]
}

func GetLogOutput(serviceName string) LogInfo {

	logInfo := LogInfo{
		Host: "10.50.10.3",
		Port: 12201,
		Type: "gelf",
	}

	hostPair, _, err := kv.Get(serviceName+"/log/host", nil)

	CheckError(err)

	if hostPair != nil {
		logInfo.Host = string(hostPair.Value)
	}

	portPair, _, err := kv.Get(serviceName+"/log/port", nil)
	CheckError(err)

	if portPair != nil {
		logInfo.Port, _ = strconv.Atoi(string(portPair.Value))
	}

	typePair, _, err := kv.Get(serviceName+"/log/type", nil)
	CheckError(err)

	if typePair != nil {
		logInfo.Type = string(typePair.Value)
	}

	return logInfo
}

func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(3)
	}
}
