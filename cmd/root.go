/*
 * Created: 2022-09-10 11:42:22
 * Author : Win-Man
 * Email : gang.shen0423@gmail.com
 * -----
 * Last Modified:
 * Modified By:
 * -----
 * Description:
 */

package cmd

import (
	"fmt"
	"os"

	"github.com/Win-Man/dbcompare/service"
	"github.com/spf13/cobra"
)

var rootCmd *cobra.Command

var (
	timeout int64
	isTest  bool
)

//sql-diff flags
var configPath string
var logLevel string
var logPath string
var sqlString string
var output string
var version bool

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// cobra.OnInitialize(initConfig)

	rootCmd = &cobra.Command{
		Use:   "dbcompare",
		Short: "dbcompare command tool",
		Long:  `A command tool for database compare`,
		RunE: func(cmd *cobra.Command, args []string) error {
			service.GetAppVersion(version)

			return nil
		},
	}

	// rootCmd.PersistentFlags().Int64Var(&timeout, "timeout", 5, "Timeout in seconds to execute")
	// rootCmd.PersistentFlags().BoolVarP(&isTest, "yes", "y", false, "run test")
	rootCmd.AddCommand(newSqlDiffCmd(), newSyncDiffCmd(), newT2OInitCmd(), newO2TInitCmd())
	rootCmd.Flags().BoolVarP(&version, "version", "V", false, "view dbcompare version")

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cobra.yaml)")
	// rootCmd.PersistentFlags().StringP("author", "a", "YOUR NAME", "author name for copyright attribution")
	// rootCmd.PersistentFlags().StringVarP(&userLicense, "license", "l", "", "name of license for the project")
	// rootCmd.PersistentFlags().Bool("viper", true, "use Viper for configuration")
	// viper.BindPFlag("author", rootCmd.PersistentFlags().Lookup("author"))
	// viper.BindPFlag("useViper", rootCmd.PersistentFlags().Lookup("viper"))
	// viper.SetDefault("author", "NAME HERE <EMAIL ADDRESS>")
	// viper.SetDefault("license", "apache")

}

func er(msg interface{}) {
	fmt.Println("Error:", msg)
	os.Exit(1)
}
