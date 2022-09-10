/*
 * Created: 2022-09-10 11:49:33
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
	"strings"

	"github.com/Win-Man/dbcompare/config"
	"github.com/Win-Man/dbcompare/pkg/logger"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)


// [] sync-diff-o2t
// [] sync-diff-inspector

func newSyncDiffCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "sync-diff",
		Short: "sync-diff",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.InitConfig(configPath)
			logger.InitLogger(logLevel, logPath, cfg)
			log.Info("Welcome to sync-diff")
			log.Debug(fmt.Sprintf("Flags:%+v", cmd.Flags()))
			log.Debug(fmt.Sprintf("arguments:%s", strings.Join(args, ",")))

			executeSqlDiff(cfg)

			return nil
		},
	}
	cmd.Flags().StringVarP(&configPath, "config", "C", "", "config file path")
	cmd.Flags().StringVarP(&logLevel, "log-level", "L", "info", "log level: info, debug, warn, error, fatal")
	cmd.Flags().StringVar(&logPath, "log-path", "", "The path of log file")
	cmd.Flags().StringVar(&sqlString, "sql", "", "single sql statement")
	cmd.Flags().StringVar(&output, "output", "", "print|file")
	return cmd
}
