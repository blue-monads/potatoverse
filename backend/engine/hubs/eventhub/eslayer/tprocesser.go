package eslayer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/blue-monads/turnix/backend/engine/hubs/eventhub/rengine"
	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
	"github.com/blue-monads/turnix/backend/utils/kosher"
	"github.com/blue-monads/turnix/backend/utils/qq"
	"github.com/blue-monads/turnix/backend/xtypes"
	"github.com/tidwall/pretty"
)

func (e *ESLayer) targetProcessor(targetId int64) error {

	qq.Println("targetProcessor/1")

	sops := e.db.GetSpaceOps()

	qq.Println("targetProcessor/2")

	target, err := e.sink.TransitionTargetStart(targetId)
	if err != nil {
		qq.Println("targetProcessor/3", err)
		return err
	}

	qq.Println("targetProcessor/4")

	event, err := e.sink.GetEvent(target.EventID)
	if err != nil {
		qq.Println("targetProcessor/5", err)
		return err
	}

	qq.Println("targetProcessor/6")

	sub, err := sops.GetEventSubscription(event.InstallID, target.SubscriptionID)
	if err != nil {
		qq.Println("targetProcessor/7", err)
		return err
	}

	qq.Println("targetProcessor/8", sub)

	ok, err := rengine.RuleEngine(sub.Rules, event.Payload)
	if err != nil {
		qq.Println("targetProcessor/9", err)
		e.sink.TransitionTargetFail(event.ID, targetId, err.Error())
		return err
	}
	if !ok {
		qq.Println("targetProcessor/10", "rules not matched")
		return nil
	}

	ectx := TargetExecution{
		Subscription: sub,
		Target:       target,
		Event:        event,
		app:          e.app,
	}

	if sub.DelayStart > 0 && target.Status != "start_delayed" {
		delayStart := time.Now().Unix() + sub.DelayStart*1000
		err = e.sink.TransitionTargetStartDelayed(targetId, event.ID, delayStart)
		if err != nil {
			qq.Println("targetProcessor/11", err)
			e.sink.TransitionTargetFail(event.ID, targetId, err.Error())
			return err
		}

		qq.Println("targetProcessor/12", "delayed for", sub.DelayStart, "seconds")
		return nil
	}

	switch sub.TargetType {
	case "webhook":
		err = PerformWebhookTargetExecution(ectx)
	case "script":
		err = PerformScriptTargetExecution(ectx)
	case "space_method":
		err = PerformSpaceMethodTargetExecution(ectx)
	case "log":
		err = PerformLogTargetExecution(ectx)
	default:
		qq.Println("targetProcessor/11", "unknown target type", sub.TargetType)
		e.sink.TransitionTargetFail(event.ID, targetId, "unknown target type: "+sub.TargetType)
		return fmt.Errorf("unknown target type: %s", sub.TargetType)
	}
	if err != nil {
		qq.Println("targetProcessor/11", err, sub.TargetType)
		e.sink.TransitionTargetFail(event.ID, targetId, err.Error())

		if sub.MaxRetries > 0 && target.RetryCount < sub.MaxRetries {
			retryCount := target.RetryCount + 1
			delay := time.Now().Unix() + sub.RetryDelay*1000
			err = e.sink.TransitionTargetDelay(targetId, event.ID, delay, retryCount)
			if err != nil {
				qq.Println("targetProcessor/12", err)
				e.sink.TransitionTargetFail(event.ID, targetId, err.Error())
				return err
			}
			qq.Println("targetProcessor/13", "delayed for", sub.RetryDelay, "seconds")
			return nil
		}

		e.sink.TransitionTargetFail(event.ID, targetId, err.Error())

		return nil
	} else {
		err = e.sink.TransitionTargetComplete(event.ID, targetId)
		if err != nil {
			qq.Println("targetProcessor/9", err)
			return err
		}

	}

	return nil

}

type TargetExecution struct {
	app          xtypes.App
	Subscription *dbmodels.MQSubscription
	Target       *dbmodels.MQEventTarget
	Event        *dbmodels.MQEvent
}

func PerformWebhookTargetExecution(ectx TargetExecution) error {
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

func PerformScriptTargetExecution(execution TargetExecution) error {
	// script := execution.SubscriptionID.TargetCode

	return nil
}

func PerformSpaceMethodTargetExecution(execution TargetExecution) error {

	return nil
}

func PerformLogTargetExecution(ectx TargetExecution) error {
	qq.Println("PerformLogTargetExecution", ectx.Event.Payload)
	result := pretty.Color(ectx.Event.Payload, nil)
	qq.Println("PerformLogTargetExecution/1", string(result))

	return nil
}
