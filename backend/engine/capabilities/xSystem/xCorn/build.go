package xcorn

import (
	"time"

	"github.com/blue-monads/potatoverse/backend/registry"
	"github.com/blue-monads/potatoverse/backend/xtypes"
	"github.com/blue-monads/potatoverse/backend/xtypes/xcapability"
	"github.com/gin-gonic/gin"
)

var (
	Name         = "xCorn"
	Icon         = `<i class="fa-solid fa-clock"></i>`
	OptionFields = []xcapability.CapabilityOptionField{}
)

func init() {

	b := xcapability.CapabilityBuilderFactory{
		Builder: func(app any) (xcapability.CapabilityBuilder, error) {
			return &CornBuilder{
				app: app.(xtypes.App),
			}, nil
		},
		Name:             Name,
		Icon:             Icon,
		FreeFieldOptions: true,
		OptionFields:     OptionFields,
	}

	registry.RegisterCapability("xcorn", b)
}

type CornBuilder struct {
	app xtypes.App
}

func (b *CornBuilder) Name() string { return Name }

func (b *CornBuilder) Build(handle xcapability.XCapabilityHandle) (xcapability.Capability, error) {
	jobs, err := loadOptions(handle)
	if err != nil {
		return nil, err
	}

	c := &CornCapability{builder: b, handle: handle, jobs: jobs}

	go c.loop()

	return c, nil
}

func (b *CornBuilder) Serve(ctx *gin.Context) {}

func (b *CornBuilder) GetDebugData() map[string]any {
	return map[string]any{}
}

// loadOptions parses free-form options where keys are job names
// and values are Go duration strings (e.g. "30s", "5m", "1h").
func loadOptions(handle xcapability.XCapabilityHandle) (map[string]*CornJob, error) {
	opts := handle.GetOptionsAsLazyData()
	if opts == nil {
		return nil, nil
	}

	optMap, err := opts.AsMap()
	if err != nil {
		return nil, err
	}

	jobs := make(map[string]*CornJob)

	for key, val := range optMap {
		strVal, ok := val.(string)
		if !ok {
			continue
		}

		dur, err := time.ParseDuration(strVal)
		if err != nil {
			continue
		}

		if existing, exists := jobs[key]; exists {
			existing.Interval = dur
		} else {
			jobs[key] = &CornJob{
				Name:     key,
				Interval: dur,
			}
		}
	}

	return jobs, nil
}
