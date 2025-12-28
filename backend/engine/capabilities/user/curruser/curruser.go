package curruser

import (
	"errors"

	"github.com/blue-monads/turnix/backend/engine/registry"
	"github.com/blue-monads/turnix/backend/services/corehub"
	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
	"github.com/blue-monads/turnix/backend/xtypes"
	"github.com/blue-monads/turnix/backend/xtypes/lazydata"
	"github.com/blue-monads/turnix/backend/xtypes/xcapability"
	"github.com/gin-gonic/gin"
)

var (
	Name         = "curruser"
	Icon         = ""
	OptionFields = []xcapability.CapabilityOptionField{}
)

func init() {
	registry.RegisterCapability(Name, xcapability.CapabilityBuilderFactory{
		Builder: func(app any) (xcapability.CapabilityBuilder, error) {
			appTyped := app.(xtypes.App)
			return &CurrUserBuilder{app: appTyped}, nil
		},
		Name:         Name,
		Icon:         Icon,
		OptionFields: OptionFields,
	})
}

type CurrUserBuilder struct {
	app xtypes.App
}

func (b *CurrUserBuilder) Build(handle xcapability.XCapabilityHandle) (xcapability.Capability, error) {
	model := handle.GetModel()
	return &CurrUserCapability{
		app:        b.app,
		handle:     handle,
		spaceId:    model.SpaceID,
		installId:  model.InstallID,
		capability: model,
	}, nil
}

func (b *CurrUserBuilder) Serve(ctx *gin.Context) {}

func (b *CurrUserBuilder) Name() string {
	return Name
}

func (b *CurrUserBuilder) GetDebugData() map[string]any {
	return map[string]any{}
}

type CurrUserCapability struct {
	app        xtypes.App
	handle     xcapability.XCapabilityHandle
	spaceId    int64
	installId  int64
	capability *dbmodels.SpaceCapability
}

func (c *CurrUserCapability) Reload(model *dbmodels.SpaceCapability) (xcapability.Capability, error) {
	c.capability = model
	return c, nil
}

func (c *CurrUserCapability) Close() error {
	return nil
}

func (c *CurrUserCapability) Handle(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"message":    "curruser capability",
		"capability": Name,
		"space_id":   c.spaceId,
	})
}

func (c *CurrUserCapability) ListActions() ([]string, error) {
	return []string{"send_user_message", "get_user_info", "get_user_config"}, nil
}

func (c *CurrUserCapability) Execute(name string, params lazydata.LazyData) (any, error) {
	switch name {
	case "send_user_message":
		return c.sendUserMessage(params)
	case "get_user_info":
		return c.getUserInfo(params)
	case "get_user_config":
		return c.getUserConfig(params)
	default:
		return nil, errors.New("unknown action: " + name)
	}
}

func (c *CurrUserCapability) parseUserToken(userToken string) (int64, error) {
	claim, err := c.app.Signer().ParseAccess(userToken)
	if err != nil {
		return 0, errors.New("invalid user_token: " + err.Error())
	}

	return claim.UserId, nil
}

type UserMessageData struct {
	UserToken   string               `json:"user_token"`
	MessageData dbmodels.UserMessage `json:"message_data"`
}

func (c *CurrUserCapability) sendUserMessage(params lazydata.LazyData) (any, error) {

	userMessageData := &UserMessageData{}

	userId, err := c.parseUserToken(userMessageData.UserToken)
	if err != nil {
		return nil, err
	}

	err = params.AsJson(&userMessageData)
	if err != nil {
		return nil, err
	}

	userMessageData.MessageData.ToUser = userId
	userMessageData.MessageData.FromUserId = 0
	userMessageData.MessageData.FromSpaceId = c.spaceId
	userMessageData.MessageData.IsRead = false

	// Use corehub to send message
	coreHub := c.app.CoreHub().(*corehub.CoreHub)
	id, err := coreHub.UserSendMessage(&userMessageData.MessageData)
	if err != nil {
		return nil, err
	}

	return id, nil
}

func (c *CurrUserCapability) getUserInfo(params lazydata.LazyData) (any, error) {
	userToken := params.GetFieldAsString("user_token")

	userId, err := c.parseUserToken(userToken)
	if err != nil {
		return nil, err
	}

	// Get user from database
	user, err := c.app.Database().GetUserOps().GetUser(userId)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (c *CurrUserCapability) getUserConfig(_ lazydata.LazyData) (any, error) {
	// userToken := params.GetFieldAsString("user_token")

	// TODO: Implement this

	return nil, nil

}
