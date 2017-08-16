package core

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

var (
	cachedRootDir    string
	cachedProjectDir string
	cachedConfig     *Config
)

type (
	Config struct {
		Env    string
		Paging struct {
			Limit    int
			MaxLimit int `yaml:"maxLimit"`
			Offset   int
		}
		Database map[string]DatabaseConfig
	}

	DatabaseConfig struct {
		Schema            string
		Adapter           string
		Charset           string
		Username          string
		Password          string
		Database          string
		Port              string
		Plural            bool
		MaxConnectionPool int
		LogMode           bool `yaml:"logmode"`
	}
)

func Check(err error) {
	if err != nil {
		panic(err)
	}
}

func GetRootDir() string {
	if cachedRootDir == "" {
		gopath := strings.Split(os.Getenv("GOPATH"), ":")[0]
		// 실제로 파일이 존재하는지 체크해야함
		cachedRootDir = path.Join(gopath, "./src/github.com/finwhale/octopus")
	}

	return cachedRootDir
}

func SetProjectDir(projectDir string) {
	cachedProjectDir = projectDir
}

func GetProjectDir() string {
	if cachedProjectDir == "" {
		dir, err := os.Getwd()
		Check(err)

		cachedProjectDir = dir
	}

	return cachedProjectDir
}

func GetConfig(reload bool) *Config {
	if reload || cachedConfig == nil {
		file, err := ioutil.ReadFile(path.Join(GetProjectDir(), ConfigFilename))

		if err != nil {
			return &Config{
				Database: map[string]DatabaseConfig{
					"test": DatabaseConfig{
						Adapter: "mysql",
						Charset: "utf8",
						Schema:  "spoon_test",
					},
				},
			}
		}

		err = yaml.Unmarshal(file, &cachedConfig)
		Check(err)
	}

	return cachedConfig
}

func CopyFolder(srcDir string, destDir string) {
	src, err := os.Stat(srcDir)
	Check(err)

	if !src.IsDir() {
		panic(fmt.Errorf("`%v` is invalid directory.", srcDir))
	}

	_, err = os.Open(destDir)
	if os.IsNotExist(err) {
		os.MkdirAll(destDir, 0777)
	}

	directory, _ := os.Open(srcDir)
	srcFileInfos, err := directory.Readdir(-1)
	Check(err)

	for _, srcFileInfo := range srcFileInfos {
		if srcFileInfo.IsDir() {
			CopyFolder(path.Join(srcDir, srcFileInfo.Name()), path.Join(destDir, srcFileInfo.Name()))
		} else {
			srcFile, err := os.Open(path.Join(srcDir, srcFileInfo.Name()))
			destFile, err := os.Create(path.Join(destDir, srcFileInfo.Name()))
			Check(err)
			_, err = io.Copy(destFile, srcFile)
			Check(err)
		}
	}
}
