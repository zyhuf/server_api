package main

import (
	"log"

	"reyzar.com/server-api/pkg/logger"
	"reyzar.com/server-api/server"
)

func main() {
	logger.InitLogFile(ServerApiFileLogName)
	logger.Debug("Starting server-api process...")
	config, err := server.GetConfig()
	logger.Debug(config)
	if err != nil {
		log.Fatal(err)
	}

	s := server.NewServer(config)
	s.Start()
}
