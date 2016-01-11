package cli

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"

	"github.com/ch3lo/overlord/configuration"
	"github.com/ch3lo/overlord/logger"
	"github.com/ch3lo/overlord/version"
	"github.com/codegangsta/cli"
)

var config *configuration.Configuration

func globalFlags() []cli.Flag {
	flags := []cli.Flag{
		cli.BoolFlag{
			Name:  "debug",
			Usage: "Debug de la app",
		},
		cli.StringFlag{
			Name:   "log-level",
			Value:  "info",
			Usage:  "Nivel de verbosidad de log",
			EnvVar: "OVERLORD_LOG_LEVEL",
		},
		cli.StringFlag{
			Name:   "log-formatter",
			Value:  "text",
			Usage:  "Formato de log",
			EnvVar: "OVERLORD_LOG_FORMATTER",
		},
		cli.BoolFlag{
			Name:   "log-colored",
			Usage:  "Coloreo de log :D",
			EnvVar: "OVERLORD_LOG_COLORED",
		},
		cli.StringFlag{
			Name:   "log-output",
			Value:  "console",
			Usage:  "Output de los logs. console | file",
			EnvVar: "OVERLORD_LOG_OUTPUT",
		},
		cli.StringFlag{
			Name:  "config",
			Value: "overlord.yaml",
			Usage: "Ruta del archivo de configuración",
		},
	}

	return flags
}

func setupConfiguration(configFile string) (*configuration.Configuration, error) {
	_, err := os.Stat(configFile)
	if os.IsNotExist(err) {
		return nil, err
	}

	configFile, err = filepath.Abs(configFile)
	if err != nil {
		return nil, err
	}

	var yamlFile []byte
	if yamlFile, err = ioutil.ReadFile(configFile); err != nil {
		return nil, err
	}

	var config configuration.Configuration
	if err = yaml.Unmarshal(yamlFile, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func setupApplication(c *cli.Context) error {
	logConfig := logger.Config{
		Level:     c.String("log-level"),
		Formatter: c.String("log-formatter"),
		Colored:   c.Bool("log-colored"),
		Output:    c.String("log-output"),
		Debug:     c.Bool("debug"),
	}

	err := logger.Configure(logConfig)
	if err != nil {
		return err
	}

	if config, err = setupConfiguration(c.String("config")); err != nil {
		return err
	}

	return nil
}

func RunApp() {
	app := cli.NewApp()
	app.Name = "overlord"
	app.Usage = "Monitor de contenedores"
	app.Version = version.VERSION + " (" + version.GITCOMMIT + ")"

	app.Flags = globalFlags()

	app.Before = func(c *cli.Context) error {
		return setupApplication(c)
	}

	app.Commands = commands

	if err := app.Run(os.Args); err != nil {
		logger.Instance().Fatalln(err)
	}
}
