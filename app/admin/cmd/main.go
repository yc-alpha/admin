package main

import (
	"os"

	"github.com/go-kratos/kratos/v2"
	"github.com/spf13/cobra"
	admin "github.com/yc-alpha/admin/app/admin"
	"github.com/yc-alpha/admin/app/admin/internal/server"
	"github.com/yc-alpha/admin/common/log_adapter"
	"github.com/yc-alpha/admin/common/snowflake"
	"github.com/yc-alpha/config"
	"github.com/yc-alpha/logger"
)

var (
	// Name is the name of the compiled software.
	Name = "paas.admin"
	// Version is the version of the compiled software.
	Version string
	// flagConf is the config flag.
	cfgFile string
	// flagRelease determines run mode.
	release bool
	// flagLog is the path of log dir.
	logFile string
	// cobra command
	rootCmd = &cobra.Command{
		Use:   "app",
		Short: Name,
		Run: func(cmd *cobra.Command, args []string) {
			logger.Info("Welcome to the Admin Application!")
			if len(args) == 0 {
				return
			}
			os.Exit(0)
		},
	}
	id, _ = os.Hostname()
	uid   = Name + "-" + id
)

func init() {
	// 注册 flag
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "conf", "c", "../configs/config.yml", "config path, eg: -c config.yaml")
	rootCmd.PersistentFlags().BoolVar(&release, "release", false, "run in release mode")
	rootCmd.PersistentFlags().StringVar(&logFile, "log", "./runtime.log", "log file path")
	// 执行 rootCmd
	if err := rootCmd.Execute(); err != nil {
		logger.Error(err)
		os.Exit(1)
	}
	for _, arg := range os.Args {
		if arg == "-h" || arg == "--help" {
			os.Exit(0) // 遇到 help 就不执行后续打印
		}
	}
	config.Load(cfgFile)
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
		kratos.Logger(log_adapter.NewAdapter()), // 使用自定义日志适配器
		// kratos.Registrar(data.NewRegistrar()),
	)
	if err := app.Run(); err != nil {
		panic(err)
	}
}
