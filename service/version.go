/*
 * Created: 2022-09-08 16:26:48
 * Author : Win-Man
 * Email : gang.shen0423@gmail.com
 * -----
 * Last Modified:
 * Modified By:
 * -----
 * Description:
 */
package service

import (
	"fmt"
	"runtime"
)

// 版本信息
var (
	Version   = "None"
	BuildTS   = "None"
	GitHash   = "None"
	GitBranch = "None"
)

func GetAppVersion(version bool) {
	if version {
		fmt.Printf("%v", getRawVersion())
	}
}

// 版本信息输出重定向到日志
//  func RecordAppVersion(app string, logger *zap.Logger, cfg *CfgFile) {
// 	 logger.Info("Welcome to "+app,
// 		 zap.String("Release Version", Version),
// 		 zap.String("Git Commit Hash", GitHash),
// 		 zap.String("Git Branch", GitBranch),
// 		 zap.String("UTC Build Time", BuildTS),
// 		 zap.String("Go Version", runtime.Version()),
// 	 )
// 	 logger.Info(app+" config", zap.Stringer("config", cfg))
//  }

func getRawVersion() string {
	info := ""
	info += fmt.Sprintf("Release Version: %s\n", Version)
	info += fmt.Sprintf("Git Commit Hash: %s\n", GitHash)
	info += fmt.Sprintf("Git Branch: %s\n", GitBranch)
	info += fmt.Sprintf("UTC Build Time: %s\n", BuildTS)
	info += fmt.Sprintf("Go Version: %s\n", runtime.Version())
	return info
}
