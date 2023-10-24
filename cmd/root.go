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
	"github.com/Win-Man/dbcompare/service"
	"github.com/spf13/cobra"
)

var rootCmd *cobra.Command

// sql-diff flags
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

	rootCmd.AddCommand(newSqlDiffCmd(), newSyncDiffCmd(), newT2OInitCmd(), newO2TInitCmd())
	rootCmd.Flags().BoolVarP(&version, "version", "V", false, "view dbcompare version")

}
