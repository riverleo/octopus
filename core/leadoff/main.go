package main

import (
	"./customs"
	"./models"
	"flag"
	"github.com/finwhale/octopus/core"
	"github.com/finwhale/octopus/server"
)

var env, port, dir string

func init() {
	flag.StringVar(&env, "env", "local", "")
	flag.StringVar(&dir, "dir", "", "")
	flag.StringVar(&port, "port", "40000", "")
	flag.Parse()
}

func main() {
	if dir != "" {
		core.SetProjectDir(dir)
	}

	models.SetUp(env)
	customs.SetUp()
	server.Run(env, port)
}
