package mysql

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wal-g/tracelog"
	"github.com/wal-g/wal-g/internal"
	"github.com/wal-g/wal-g/internal/databases/mysql"
	"github.com/wal-g/wal-g/utility"
)

const (
	xtrabackupPushShortDescription = "Creates new backup and pushes it to storage"

	fullBackupFlag        = "full"
	deltaFromUserDataFlag = "delta-from-user-data"
	deltaFromNameFlag     = "delta-from-name"

	fullBackupShorthand = "f"
)

var (
	// backupPushCmd represents the streamPush command
	xtrabackupPushCmd = &cobra.Command{
		Use:   "xtrabackup-push",
		Short: xtrabackupPushShortDescription,
		PreRun: func(cmd *cobra.Command, args []string) {
			internal.RequiredSettings[internal.NameStreamCreateCmd] = true
			internal.RequiredSettings[internal.MysqlDatasourceNameSetting] = true
			err := internal.AssertRequiredSettingsSet()
			tracelog.ErrorLogger.FatalOnError(err)
		},
		Run: func(cmd *cobra.Command, args []string) {
			internal.ConfigureLimiters()

			// FIXME: do we need this?
			if deltaFromName == "" {
				deltaFromName = viper.GetString(internal.DeltaFromNameSetting)
			}
			if deltaFromUserData == "" {
				deltaFromUserData = viper.GetString(internal.DeltaFromUserDataSetting)
			}

			deltaBaseSelector, err := internal.NewDeltaBaseSelector(
				deltaFromName, deltaFromUserData, mysql.NewGenericMetaFetcher())
			tracelog.ErrorLogger.FatalOnError(err)

			uploader, err := internal.ConfigureSplitUploader()
			tracelog.ErrorLogger.FatalOnError(err)
			folder := uploader.Folder()
			uploader.ChangeDirectory(utility.BaseBackupPath)
			backupCmd, err := internal.GetCommandSetting(internal.NameStreamCreateCmd)
			tracelog.ErrorLogger.FatalOnError(err)

			if userData == "" {
				userData = viper.GetString(internal.SentinelUserDataSetting)
			}

			mysql.HandleBackupPush(
				folder,
				uploader,
				backupCmd,
				permanent,
				fullBackup,
				userData,
				mysql.NewRegularDeltaBackupConfigurator(folder, deltaBaseSelector),
			)
		},
	}
	fullBackup        = true
	deltaFromName     = ""
	deltaFromUserData = ""
)

func init() {
	cmd.AddCommand(xtrabackupPushCmd)

	// TODO: Merge similar backup-push functionality
	// to avoid code duplication in command handlers
	xtrabackupPushCmd.Flags().BoolVarP(&permanent, permanentFlag, permanentShorthand,
		false, "Pushes permanent backup")
	xtrabackupPushCmd.Flags().BoolVarP(&fullBackup, fullBackupFlag, fullBackupShorthand,
		true, "Make full backup-push")
	xtrabackupPushCmd.Flags().StringVar(&deltaFromName, deltaFromNameFlag,
		"", "Select the backup specified by name as the target for the delta backup")
	xtrabackupPushCmd.Flags().StringVar(&deltaFromUserData, deltaFromUserDataFlag,
		"", "Select the backup specified by UserData as the target for the delta backup")
	xtrabackupPushCmd.Flags().StringVar(&userData, addUserDataFlag,
		"", "Write the provided user data to the backup sentinel and metadata files.")
}
