package core

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func Install() {
	Run("go get github.com/go-sql-driver/mysql")
	Run("go get gopkg.in/yaml.v2")
	Run("go get github.com/labstack/echo")
	Run("go get github.com/labstack/echo/middleware")
}

func Run(command string) {
	fmt.Println(command)
	args := strings.Split(command, " ")
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("command failed: %s\n", command)
		panic(err)
	}
}
