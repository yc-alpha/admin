package main

import (
	"flag"
	"os"

	"github.com/go-kratos/kratos/v2"
	admin "github.com/yc-alpha/admin/app/admin"
	"github.com/yc-alpha/admin/app/admin/internal/server"
	"github.com/yc-alpha/admin/common/snowflake"
	"github.com/yc-alpha/config"
)

var (
	// Name is the name of the compiled software.
	Name = "paas.user_management"
	// Version is the version of the compiled software.
	Version string
	// flagConf is the config flag.
	flagConf string
	// flagRelease determines run mode.
	flagRelease bool
	// flagLog is the path of log dir.
	flagLog string
	id, _   = os.Hostname()
	uid     = Name + "-" + id
)

func init() {
	flag.StringVar(&flagConf, "conf", "../configs/config.yml", "config path, eg: -conf config.yaml")
	flag.BoolVar(&flagRelease, "release", false, "run mode, eg: -release true")
	flag.StringVar(&flagLog, "log", "./runtime.log", "log dir, eg: -log ./runtime.log")
	config.Load(flagConf)

	snowflake.SetNode(config.Get("node").ToInt64())
}

func main() {

	httpServer := server.NewHTTPServer()
	grpcServer := server.NewGRPCServer()
	admin.RegisteApplication(httpServer, grpcServer)
	app := kratos.New(
		kratos.ID(uid),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Server(httpServer, grpcServer),
		// kratos.Registrar(data.NewRegistrar()),
	)

	if err := app.Run(); err != nil {
		panic(err)
	}
}
