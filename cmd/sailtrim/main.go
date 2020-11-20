package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/alecthomas/kingpin"
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

	kingpin.Command("deploy", "create new deployment")
	kingpin.Command("update", "update container service")
	dump := kingpin.Command("dump", "dump container service")
	dumpService := dump.Flag("name", "service name").Required().String()

	command := kingpin.Parse()
	if command == "version" {
		fmt.Println("sailtrim", Version)
		return 0
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

	log.Println("[info] sailtrim", Version)
	switch command {
	case "deploy":
		err = app.Deploy(ctx)
	case "update":
		err = app.Update(ctx)
	case "dump":
		err = app.Dump(ctx, *dumpService)
	}

	if err != nil {
		log.Println("[error]", err)
		return 1
	}
	log.Println("[info] completed")
	return 0
}
