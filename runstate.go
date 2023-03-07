package qcli

type RunState int

const (
	RunStateDebug RunState = iota
	RunStateFinishMigrate
	RunStateInMigrate
	RunStateInternalError
	RunStateIOError
	RunStatePaused
	RunStatePostMigrate
	RunStatePreLaunch
	RunStateRestoreVM
	RunStateRunning
	RunStateSaveVM
	RunStateShutdown
	RunStateSuspended
	RunStateWatchdog
	RunStateGuestPanicked
	RunStateColo
	RunStateUnknown
)

const (
	RunStateDebugStr         string = "debug"
	RunStateFinishMigrateStr string = "finish-migrate"
	RunStateInMigrateStr     string = "inmigrate"
	RunStateInternalErrorStr string = "internal-error"
	RunStateIOErrorStr       string = "io-error"
	RunStatePausedStr        string = "paused"
	RunStatePostMigrateStr   string = "postmigrate"
	RunStatePreLaunchStr     string = "prelaunch"
	RunStateRestoreVMStr     string = "restore-vm"
	RunStateRunningStr       string = "running"
	RunStateSaveVMStr        string = "save-vm"
	RunStateShutdownStr      string = "shutdown"
	RunStateSuspendedStr     string = "suspended"
	RunStateWatchdogStr      string = "watchdog"
	RunStateGuestPanickedStr string = "guest-panicked"
	RunStateColoStr          string = "colo"
	RunStateUnknownStr       string = "unknown"
)

func (rs RunState) String() string {
	toString := map[RunState]string{
		RunStateDebug:         RunStateDebugStr,
		RunStateFinishMigrate: RunStateFinishMigrateStr,
		RunStateInMigrate:     RunStateInMigrateStr,
		RunStateInternalError: RunStateInternalErrorStr,
		RunStateIOError:       RunStateIOErrorStr,
		RunStatePaused:        RunStatePausedStr,
		RunStatePostMigrate:   RunStatePostMigrateStr,
		RunStatePreLaunch:     RunStatePreLaunchStr,
		RunStateRestoreVM:     RunStateRestoreVMStr,
		RunStateRunning:       RunStateRunningStr,
		RunStateSaveVM:        RunStateSaveVMStr,
		RunStateShutdown:      RunStateShutdownStr,
		RunStateSuspended:     RunStateSuspendedStr,
		RunStateWatchdog:      RunStateWatchdogStr,
		RunStateGuestPanicked: RunStateGuestPanickedStr,
		RunStateColo:          RunStateColoStr,
		RunStateUnknown:       RunStateUnknownStr,
	}

	if rs, ok := toString[rs]; ok {
		return rs
	}
	return RunStateUnknownStr
}

func ToRunState(status string) RunState {
	toRunState := map[string]RunState{
		RunStateDebugStr:         RunStateDebug,
		RunStateFinishMigrateStr: RunStateFinishMigrate,
		RunStateInMigrateStr:     RunStateInMigrate,
		RunStateInternalErrorStr: RunStateInternalError,
		RunStateIOErrorStr:       RunStateIOError,
		RunStatePausedStr:        RunStatePaused,
		RunStatePostMigrateStr:   RunStatePostMigrate,
		RunStatePreLaunchStr:     RunStatePreLaunch,
		RunStateRestoreVMStr:     RunStateRestoreVM,
		RunStateRunningStr:       RunStateRunning,
		RunStateSaveVMStr:        RunStateSaveVM,
		RunStateShutdownStr:      RunStateShutdown,
		RunStateSuspendedStr:     RunStateSuspended,
		RunStateWatchdogStr:      RunStateWatchdog,
		RunStateGuestPanickedStr: RunStateGuestPanicked,
		RunStateColoStr:          RunStateColo,
		RunStateUnknownStr:       RunStateUnknown,
	}

	if rs, ok := toRunState[status]; ok {
		return rs
	}

	return RunStateUnknown
}
