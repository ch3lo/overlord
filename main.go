package main

import (
	"github.com/ch3lo/overlord/cli"
	//Necesarios para que funcione el init()
	_ "github.com/ch3lo/overlord/notification/email"
	_ "github.com/ch3lo/overlord/notification/http"
	_ "github.com/latam-airlines/mesos-framework-factory/marathon"
	_ "github.com/latam-airlines/mesos-framework-factory/swarm"
)

func main() {
	cli.RunApp()
}
