package eventhub

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
	"github.com/blue-monads/turnix/backend/utils/kosher"
	"github.com/blue-monads/turnix/backend/utils/qq"
	"github.com/tidwall/gjson"
	"github.com/tidwall/pretty"
)

func (e *EventHub) targetProcessor(targetId int64) error {

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

	ok, err := RuleEngine(sub.Rules, event.Payload)
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
		return err
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
	Subscription *dbmodels.EventSubscription
	Target       *dbmodels.MQEventTarget
	Event        *dbmodels.MQEvent
}

func RuleEngine(rulestr string, payload []byte) (bool, error) {

	if rulestr == "{}" || rulestr == "" {
		return true, nil
	}

	rules := map[string]string{}
	err := json.Unmarshal([]byte(rulestr), &rules)
	if err != nil {
		return false, err
	}

	json := kosher.Str(payload)
	for k, v := range rules {
		qq.Println("RuleEngine/1", k, v)
		value := gjson.Get(json, k)
		if value.String() != v {
			qq.Println("RuleEngine/2", value.String(), v)
			return false, nil
		}
		qq.Println("RuleEngine/3", value.String(), v)
	}

	return true, nil
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
