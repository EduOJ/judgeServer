package main

func Serve() {
	initConsoleLogger()
	readConfig()
	initFileLogger()
	initHttpClient()
}
