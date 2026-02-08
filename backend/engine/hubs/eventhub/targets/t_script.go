package eslayer

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/blue-monads/potatoverse/backend/engine/executors/luaz"
	"github.com/blue-monads/potatoverse/backend/engine/hubs/eventhub/evtype"
	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/blue-monads/potatoverse/backend/utils/qq"
	"github.com/blue-monads/potatoverse/backend/xtypes"
	"github.com/blue-monads/potatoverse/backend/xtypes/lazydata"
)

type Exec struct {
	inner     xtypes.Executor
	createdAt int64
}

func PerformScriptTargetExecution(app xtypes.App) func(execution *evtype.TExecution) error {

	sops := app.Database().GetSpaceOps()
	pkgOps := app.Database().GetPackageInstallOps()

	builder, err := luaz.BuildLuazExecutorBuilder(app)
	if err != nil {
		panic(err)
	}

	// Lru grabage Collect not used
	executors := make(map[string]*Exec)
	elock := sync.RWMutex{}

	logger := app.Logger().With("submod", "target/exec")

	cleanup := func() {
		// Clean up executors older than 1 hour
		cutoff := time.Now().Add(-1 * time.Hour)

		elock.Lock()
		defer elock.Unlock()

		for key, execEntry := range executors {
			if time.Unix(0, execEntry.createdAt).Before(cutoff) {
				execEntry.inner.Cleanup()
				delete(executors, key)
				logger.Debug("cleaned up old executor", "key", key)
			}
		}
	}

	getExecutor := func(sub *dbmodels.MQSubscription) *Exec {
		elock.RLock()
		exe, found := executors[sub.TargetCode]
		elock.RUnlock()

		if found {
			return exe
		}

		elock.Lock()
		defer elock.Unlock()

		// Double-check after acquiring write lock
		exe, found = executors[sub.TargetCode]
		if found {
			return exe
		}

		space, err := sops.GetSpace(sub.SpaceID)
		if err != nil {
			return nil
		}

		pkg, err := pkgOps.GetPackage(space.InstalledId)
		if err != nil {
			return nil
		}

		iroot, err := os.OpenRoot("test")
		if err != nil {
			return nil
		}

		// Build new executor
		ie, err := builder.Build(&xtypes.ExecutorBuilderOption{
			SpaceId:          sub.SpaceID,
			WorkingFolder:    "",
			Logger:           logger,
			PackageVersionId: pkg.ActiveInstallID,
			CodeLoader: func() (string, error) {
				return sub.TargetCode, nil
			},
			InstalledId: pkg.ID,
			FsRoot:      iroot,
		})

		if err != nil {
			qq.Println("@script_executor_build_error", err)
			return nil
		}

		exe = &Exec{
			inner:     ie,
			createdAt: time.Now().UnixNano(),
		}

		executors[sub.TargetCode] = exe
		return exe
	}

	return func(execution *evtype.TExecution) error {
		subid := execution.Subscription.ID
		spaceId := execution.Subscription.TargetSpaceID

		qq.Println("@script_target_execution", subid, spaceId, execution.Event.Name)

		// Cleanup old executors
		cleanup()

		executor := getExecutor(execution.Subscription)
		if executor == nil {
			qq.Println("@script_executor_not_found", spaceId)
			return nil // Return nil to avoid retries for permanent errors
		}

		actionEvent := &xtypes.ActionEvent{
			EventType:  "event_target",
			ActionName: "on_event",
			Params: map[string]string{
				"event_name": execution.Event.Name,
				"event_id":   fmt.Sprint(execution.Event.ID),
			},
			Request: nil,
		}

		err := executor.inner.HandleAction(actionEvent)
		if err != nil {
			qq.Println("@script_target_execution_error", subid, err)
			return err
		}

		qq.Println("@script_target_execution_success", subid)
		return nil
	}
}

type TScriptAction struct {
	execution *evtype.TExecution
}

func (ts *TScriptAction) ListActions() ([]string, error) { return []string{"set_retry"}, nil }

func (ts *TScriptAction) ExecuteAction(name string, params lazydata.LazyData) (any, error) {
	if name != "set_retry" {
		return nil, fmt.Errorf("Unknown action")
	}

	ts.execution.RetryAble = params.GetFieldAsBool("value")

	return nil, nil
}
