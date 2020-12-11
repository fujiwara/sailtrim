package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/alecthomas/kingpin"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/fujiwara/sailtrim"
	"github.com/hashicorp/logutils"
)

// Version number
var Version = "current"

func main() {
	os.Exit(_main())
}

func _main() int {
	ctx := context.Background()

	kingpin.Command("version", "show version")
	logLevel := kingpin.Flag("log-level", "log level (trace, debug, info, warn, error)").Default("info").Enum("trace", "debug", "info", "warn", "error")
	configPath := kingpin.Flag("config", "configuration file path").Default("config.yaml").String()
	debug := kingpin.Flag("debug", "set --log-level to debug").Bool()

	kingpin.Command("deploy", "create new deployment")
	kingpin.Command("update", "update container service")
	status := kingpin.Command("status", "show container service status")
	statusDetail := status.Flag("detail", "show full status as JSON format").Default("false").Bool()

	init := kingpin.Command("init", "initialize a container service")
	initServiceName := init.Flag("service-name", "service name").Required().String()

	logs := kingpin.Command("logs", "show logs")
	logsOpt := sailtrim.LogsOption{}
	logsOpt.ContainerName = logs.Flag("container-name", "container name").String()
	logsOpt.FilterPattern = logs.Flag("filter-pattern", "filter pattern").String()
	logsOpt.StartTimeStr = logs.Flag("start-time", "start time").String()
	logsOpt.EndTimeStr = logs.Flag("end-time", "end time").String()

	command := kingpin.Parse()
	if command == "version" {
		fmt.Println("sailtrim", Version)
		return 0
	}

	if *debug {
		logLevel = aws.String("debug")
	}
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"trace", "debug", "info", "warn", "error"},
		MinLevel: logutils.LogLevel(*logLevel),
		Writer:   os.Stderr,
	}
	log.SetOutput(filter)

	app, err := sailtrim.New(session.Must(session.NewSession()), *configPath)
	if err != nil {
		log.Println("[error]", err)
		return 1
	}

	log.Println("[debug] sailtrim", Version)
	switch command {
	case "deploy":
		err = app.Deploy(ctx)
	case "update":
		err = app.Update(ctx)
	case "init":
		err = app.Init(ctx, *initServiceName)
	case "status":
		err = app.Status(ctx, sailtrim.StatusOption{Detail: *statusDetail})
	case "logs":
		err = app.Logs(ctx, logsOpt)
	}

	if err != nil {
		log.Println("[error]", err)
		return 1
	}
	log.Println("[debug] completed")
	return 0
}
