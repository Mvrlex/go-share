package main

import (
	"errors"
	"log"
	"majo-tech.com/share/environment"
	"majo-tech.com/share/storage"
	"majo-tech.com/share/web"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	// logging
	temp, err := os.CreateTemp("", "go-share.*.log")
	if err != nil {
		log.Fatalln(errors.Join(errors.New("could not create log file"), err))
	}
	defer os.Remove(temp.Name())
	log.SetOutput(temp)
	println("initialized log file at", temp.Name())

	// storage
	storagePath := "./shared"
	storage, err := storage.NewPhysicalStorage(storagePath)
	if err != nil {
		log.Fatalln("could not create storage,", err)
	}
	defer storage.Close()
	log.Printf("storage initialized in directory %q", storagePath)

	// graceful shutdown
	shutdownSignal := make(chan os.Signal)
	signal.Notify(shutdownSignal, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-shutdownSignal
		storage.Close()
		os.Remove(temp.Name())
		os.Exit(1)
	}()

	// environment vars
	maxFileSize, err := environment.GetMaxFileSize()
	if err != nil {
		log.Fatalln("could not read max file size from environment:", err)
	}
	diskSpace, err := environment.GetDiskSpace()
	if err != nil {
		log.Fatalln("could not read allowed disk space from environment:", err)
	}

	// http server
	server := web.Server{
		Storage:          storage,
		MaxFileSizeBytes: environment.ValueOrDefault(maxFileSize, 104857600), // 100 MiB
		DiskSpaceBytes:   environment.ValueOrDefault(diskSpace, 32212254720), // 30 GiB
		Host:             environment.ValueOrDefault(environment.GetHost(), "http://localhost:8080"),
	}
	log.Fatalln(server.Start())

}
