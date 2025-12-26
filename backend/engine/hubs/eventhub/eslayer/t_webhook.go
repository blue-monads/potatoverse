package eslayer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
	"github.com/blue-monads/turnix/backend/utils/kosher"
	"github.com/blue-monads/turnix/backend/xtypes"
)

type TargetExecution struct {
	App          xtypes.App
	Subscription *dbmodels.MQSubscription
	Target       *dbmodels.MQEventTarget
	Event        *dbmodels.MQEvent
	RetryAble    bool
}

func PerformWebhookTargetExecution(ectx *TargetExecution) error {
	url := ectx.Subscription.TargetEndpoint

	bodyRaw := ectx.Event.Payload
	rbody := bytes.NewReader(bodyRaw)

	req, err := http.NewRequest(http.MethodPost, url, rbody)
	if err != nil {
		return err
	}

	if ectx.Subscription.TargetOptions != "{}" && ectx.Subscription.TargetOptions != "" {
		targetOptions := map[string]any{}
		topts := ectx.Subscription.TargetOptions
		err = json.Unmarshal(kosher.Byte(topts), &targetOptions)
		if err != nil {
			return err
		}

		for k, v := range targetOptions {
			if after, ok := strings.CutPrefix(k, "Header-"); ok {
				req.Header.Set(after, fmt.Sprintf("%v", v))
				continue
			}
		}
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("webhook failed with status code %d", resp.StatusCode)
	}

	return nil
}
