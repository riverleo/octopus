package main

import (
	"flag"
	"fmt"
	"github.com/finwhale/octopus/core"
	"path"
)

var (
	project, env       string
	isBuild, isInstall bool
)

// 커맨드 라인을 통해 넘겨받은 매개변수들을 초기화
func init() {
	flag.StringVar(&env, "env", "local", fmt.Sprintf("Choose the env defined in the %v", core.ConfigFilename))
	flag.StringVar(&project, "init", "", "Create a new octopus project")
	flag.BoolVar(&isBuild, "build", false, fmt.Sprintf("Create %v, %v", core.DBFilename, core.ModelFilename))
	flag.BoolVar(&isInstall, "install", false, fmt.Sprintf("Install dependencies"))
	flag.Parse()
}

func main() {
	if project != "" {
		core.CopyFolder(path.Join(core.GetRootDir(), "core/leadoff"), project)

		// success message
		fmt.Printf("BUILD SUCCESS!\n")
		fmt.Printf("Check the database environment in the %v and run build. (ex: octopus --build --env=local)\n", core.ConfigFilename)
	} else if isBuild {
		adapter, dbUrl, schemaName, charset, _, _, _ := core.GetSchemaInfo(env, true)
		core.Build(true, env, adapter, dbUrl, schemaName, charset)
	} else if isInstall {
		core.Install()
	} else {
		flag.PrintDefaults()
	}
}
