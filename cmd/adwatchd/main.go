package main

import (
	"context"
	"os"

	"github.com/ubuntu/adsys/cmd/adwatchd/commands"
	log "github.com/ubuntu/adsys/internal/grpc/logstreamer"

	"github.com/ubuntu/adsys/internal/consts"
	"github.com/ubuntu/adsys/internal/i18n"
)

func run(a *commands.App) int {
	i18n.InitI18nDomain(consts.TEXTDOMAIN)
	//TODO: defer installSignalHandler(a)()

	// log.SetFormatter(&log.TextFormatter{
	// 	DisableLevelTruncation: true,
	// 	DisableTimestamp:       true,
	// })

	if err := a.Run(); err != nil {
		log.Error(context.Background(), err)

		if a.UsageError() {
			return 2
		}
		return 1
	}

	return 0
}

func main() {
	app := commands.New()
	os.Exit(run(app))
}
