package main

import (
	"github.com/nange/gospider"
	_ "github.com/nange/gospider/_example/rule/baidunews"
	_ "github.com/nange/gospider/_example/rule/dianpindesigner"
	_ "github.com/nange/gospider/_example/rule/dianping"
	_ "github.com/nange/gospider/_example/rule/mojitianqi"
	_ "github.com/nange/gospider/_example/rule/stackoverflow"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{TimestampFormat: "2006-01-02 15:04:05.000"})
	log.SetLevel(log.DebugLevel)
}

func main() {
	gs := gospider.New()
	log.Fatal(gs.Run())
}
