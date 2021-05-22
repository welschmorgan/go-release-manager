package root

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/welschmorgan/go-project-manager/cmd/config"
	initCommand "github.com/welschmorgan/go-project-manager/cmd/init"
	releaseCommand "github.com/welschmorgan/go-project-manager/cmd/release"
	"gopkg.in/yaml.v2"
)

var Command = &cobra.Command{
	Use:   "grlm [commands]",
	Short: "Release multiple projects in a single go",
	Long:  `GRLM allows releasing multiple projects declared in a workspace`,
}

// Execute executes the root command.
func Execute() error {
	return Command.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	if cwd, err := os.Getwd(); err != nil {
		panic(err.Error())
	} else {
		config.Get().WorkingDirectory = cwd
	}
	// config file
	Command.PersistentFlags().StringVarP(&config.Get().CfgFile, "config", "c", config.Get().CfgFile, "config file (default is $HOME/.grlm.yaml)")
	viper.BindPFlag("config", Command.PersistentFlags().Lookup("config"))

	// verbose
	Command.PersistentFlags().BoolVarP(&config.Get().Verbose, "verbose", "v", config.Get().Verbose, "show additionnal log messages")
	viper.BindPFlag("verbose", Command.PersistentFlags().Lookup("verbose"))

	// change working dir
	Command.PersistentFlags().StringVarP(&config.Get().WorkingDirectory, "working-directory", "C", config.Get().WorkingDirectory, "change working directory")
	viper.BindPFlag("working-directory", Command.PersistentFlags().Lookup("working-directory"))

	// define workspaces root
	Command.PersistentFlags().StringVar(&config.Get().WorkspacesRoot, "workspaces-root", config.Get().WorkspacesRoot, "The root folder where to find workspaces")
	viper.BindPFlag("workspaces_root", Command.PersistentFlags().Lookup("workspaces-root"))

	// Command.ActionAddCommand(addCmd)
	Command.AddCommand(initCommand.Command)
	Command.AddCommand(releaseCommand.Command)
}

func initConfig() {
	if config.Get().CfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(config.Get().CfgFile)
	} else {
		viper.SetConfigName("grlm")        // name of config file (without extension)
		viper.SetConfigType("yaml")        // REQUIRED if the config file does not have the extension in the name
		viper.AddConfigPath("/etc/grlm/")  // path to look for the config file in
		viper.AddConfigPath("$HOME/.grlm") // call multiple times to add many search paths
		viper.AddConfigPath(".")           // optionally look for config in the working directory
	}

	viper.AutomaticEnv()

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("Config file changed:", e.Name)
	})

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
		} else {
			// Config file was found but another error was produced
			panic(fmt.Errorf("configuration error: %s", err))
		}
	}
	if config.Get().Verbose {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
	if _, err := os.Stat(config.Get().WorkingDirectory); err != nil && os.IsNotExist(err) {
		if err := os.MkdirAll(config.Get().WorkingDirectory, 0755); err != nil {
			fmt.Printf("error: %s\n", err.Error())
		}
	}
	if err := os.Chdir(config.Get().WorkingDirectory); err != nil {
		panic(err.Error())
	}
	config.WorkspacePath = filepath.Join(config.WorkingDirectory, config.WorkspaceFilename)
	if content, err := os.ReadFile(config.WorkspacePath); err != nil {
		panic(err.Error())
	} else if err = yaml.Unmarshal(content, &config.Workspace); err != nil {
		panic(err.Error())
	}
}
