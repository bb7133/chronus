package main

import (
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/angopher/chronus/raftmeta"
	imeta "github.com/angopher/chronus/services/meta"
	"github.com/angopher/chronus/x"
	"github.com/influxdata/influxdb/logger"
	"github.com/influxdata/influxdb/services/meta"
	"os"
)

func main() {
	configFile := flag.String("config", "", "-config config_file")
	flag.Parse()

	config := raftmeta.NewConfig()
	if *configFile != "" {
		x.Check((&config).FromTomlFile(*configFile))
	} else {
		toml.NewEncoder(os.Stdout).Encode(&config)
		return
	}

	fmt.Printf("config:%+v\n", config)

	metaCli := imeta.NewClient(&meta.Config{
		RetentionAutoCreate: config.RetentionAutoCreate,
		LoggingEnabled:      true,
	})
	err := metaCli.Open()
	x.Check(err)

	log := logger.New(os.Stderr)

	node := raftmeta.NewRaftNode(config)
	node.MetaCli = metaCli
	node.WithLogger(log)

	t := raftmeta.NewTransport()
	t.WithLogger(log)

	node.Transport = t
	node.InitAndStartNode()
	go node.Run()

	linearRead := raftmeta.NewLinearizabler(node)
	go linearRead.ReadLoop()

	service := raftmeta.NewMetaService(metaCli, node, linearRead)
	service.WithLogger(log)
	service.Start(config.MyAddr)
}