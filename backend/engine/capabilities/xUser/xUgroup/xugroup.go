package xugroup

import (
	"errors"

	"github.com/blue-monads/potatoverse/backend/services/corehub"
	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/blue-monads/potatoverse/backend/xtypes"
	"github.com/blue-monads/potatoverse/backend/xtypes/lazydata"
	"github.com/blue-monads/potatoverse/backend/xtypes/xcapability"
	"github.com/gin-gonic/gin"
	"github.com/upper/db/v4"
)

type UgroupOptions struct {
	AllUserGroup bool     `json:"all_user_group"`
	Ugroups      []string `json:"ugroups"`
}

type UgroupCapability struct {
	app          xtypes.App
	handle       xcapability.XCapabilityHandle
	spaceId      int64
	installId    int64
	allUserGroup bool
	ugroups      []string
}

func (c *UgroupCapability) Reload(model *dbmodels.SpaceCapability) (xcapability.Capability, error) {
	opts := &UgroupOptions{}
	c.handle.GetOptions(opts)
	c.allUserGroup = opts.AllUserGroup
	c.ugroups = opts.Ugroups
	return c, nil
}

func (c *UgroupCapability) Close() error {
	return nil
}

func (c *UgroupCapability) Handle(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"message":    "ugroup capability",
		"capability": Name,
		"space_id":   c.spaceId,
	})
}

func (c *UgroupCapability) ListActions() ([]string, error) {
	return []string{
		"query_group_users",
		"get_user_info",
		"query_user_config",
		"update_user_config",
		"delete_user_config",
		"send_user_message",
	}, nil
}

func (c *UgroupCapability) Execute(name string, params lazydata.LazyData) (any, error) {
	switch name {
	case "query_group_users":
		return c.queryGroupUsers(params)
	case "get_user_info":
		return c.getUserInfo(params)
	case "query_user_config":
		return c.queryUserConfig(params)
	case "update_user_config":
		return c.updateUserConfig(params)
	case "delete_user_config":
		return c.deleteUserConfig(params)
	case "send_user_message":
		return c.sendUserMessage(params)
	default:
		return nil, errors.New("unknown action: " + name)
	}
}

func (c *UgroupCapability) isAllowedGroup(group string) bool {
	if c.allUserGroup {
		return true
	}
	for _, g := range c.ugroups {
		if g == group {
			return true
		}
	}
	return false
}

func (c *UgroupCapability) checkUserAccess(userId int64) (*dbmodels.User, error) {
	user, err := c.app.Database().GetUserOps().GetUser(userId)
	if err != nil {
		return nil, err
	}
	if !c.isAllowedGroup(user.Ugroup) {
		return nil, errors.New("access denied: user not in allowed group")
	}
	return user, nil
}

// action params

type queryGroupUsersParams struct {
	Offset int    `json:"offset"`
	Limit  int    `json:"limit"`
	Group  string `json:"group"`
}

type queryUserConfigParams struct {
	UserID      int64  `json:"user_id"`
	ConfigGroup string `json:"config_group"`
	Limit       int    `json:"limit"`
	Offset      int    `json:"offset"`
}

type updateUserConfigParams struct {
	UserID      int64  `json:"user_id"`
	ConfigGroup string `json:"config_group"`
	ConfigKey   string `json:"config_key"`
	ConfigValue string `json:"config_value"`
}

type deleteUserConfigParams struct {
	UserID      int64  `json:"user_id"`
	ConfigGroup string `json:"config_group"`
	ConfigKey   string `json:"config_key"`
}

type sendUserMessageParams struct {
	UserID  int64  `json:"user_id"`
	Title   string `json:"title"`
	Message string `json:"message"`
}

type userConfig struct {
	ID     int64  `json:"id" db:"id,omitempty"`
	Key    string `json:"key" db:"key"`
	Group  string `json:"group" db:"group"`
	Value  string `json:"value" db:"value"`
	UserID int64  `json:"user_id" db:"user_id"`
}

// actions

func (c *UgroupCapability) queryGroupUsers(params lazydata.LazyData) (any, error) {
	var p queryGroupUsersParams
	if err := params.AsJson(&p); err != nil {
		return nil, err
	}

	if p.Limit <= 0 || p.Limit > 1000 {
		p.Limit = 100
	}

	if p.Group == "" && !c.allUserGroup {
		return nil, errors.New("group is required when all_user_group is false")
	}

	if p.Group != "" && !c.isAllowedGroup(p.Group) {
		return nil, errors.New("access denied for group: " + p.Group)
	}

	cond := map[any]any{}
	if p.Group != "" {
		cond["ugroup"] = p.Group
	}

	users, err := c.app.Database().GetUserOps().ListUserByCond(cond, p.Offset, p.Limit)
	if err != nil {
		return nil, err
	}

	for i := range users {
		users[i].Password = ""
	}

	if users == nil {
		return []dbmodels.User{}, nil
	}

	return users, nil
}

func (c *UgroupCapability) getUserInfo(params lazydata.LazyData) (any, error) {
	userId := int64(params.GetFieldAsInt("user_id"))

	user, err := c.checkUserAccess(userId)
	if err != nil {
		return nil, err
	}

	user.Password = ""
	return user, nil
}

func (c *UgroupCapability) queryUserConfig(params lazydata.LazyData) (any, error) {
	var p queryUserConfigParams
	if err := params.AsJson(&p); err != nil {
		return nil, err
	}

	if _, err := c.checkUserAccess(p.UserID); err != nil {
		return nil, err
	}

	if p.Limit <= 0 || p.Limit > 1000 {
		p.Limit = 100
	}

	configs := make([]userConfig, 0)
	cond := db.Cond{"user_id": p.UserID}
	if p.ConfigGroup != "" {
		cond["group"] = p.ConfigGroup
	}

	err := c.app.Database().Table("UserConfig").
		Find(cond).
		Offset(p.Offset).
		Limit(p.Limit).
		All(&configs)
	if err != nil {
		if c.app.Database().IsEmptyRowsError(err) {
			return []userConfig{}, nil
		}
		return nil, err
	}

	return configs, nil
}

func (c *UgroupCapability) updateUserConfig(params lazydata.LazyData) (any, error) {
	var p updateUserConfigParams
	if err := params.AsJson(&p); err != nil {
		return nil, err
	}

	if _, err := c.checkUserAccess(p.UserID); err != nil {
		return nil, err
	}

	if p.ConfigKey == "" {
		return nil, errors.New("config_key is required")
	}

	existing := &userConfig{}
	err := c.app.Database().Table("UserConfig").
		Find(db.Cond{"user_id": p.UserID, "group": p.ConfigGroup, "key": p.ConfigKey}).
		One(existing)

	if err != nil {
		if !c.app.Database().IsEmptyRowsError(err) {
			return nil, err
		}

		_, err = c.app.Database().Table("UserConfig").Insert(&userConfig{
			UserID: p.UserID,
			Group:  p.ConfigGroup,
			Key:    p.ConfigKey,
			Value:  p.ConfigValue,
		})
		if err != nil {
			return nil, err
		}
	} else {
		err = c.app.Database().Table("UserConfig").
			Find(db.Cond{"id": existing.ID}).
			Update(map[string]any{"value": p.ConfigValue})
		if err != nil {
			return nil, err
		}
	}

	return map[string]bool{"success": true}, nil
}

func (c *UgroupCapability) deleteUserConfig(params lazydata.LazyData) (any, error) {
	var p deleteUserConfigParams
	if err := params.AsJson(&p); err != nil {
		return nil, err
	}

	if _, err := c.checkUserAccess(p.UserID); err != nil {
		return nil, err
	}

	err := c.app.Database().Table("UserConfig").
		Find(db.Cond{"user_id": p.UserID, "group": p.ConfigGroup, "key": p.ConfigKey}).
		Delete()
	if err != nil {
		return nil, err
	}

	return map[string]bool{"success": true}, nil
}

func (c *UgroupCapability) sendUserMessage(params lazydata.LazyData) (any, error) {
	var p sendUserMessageParams
	if err := params.AsJson(&p); err != nil {
		return nil, err
	}

	if _, err := c.checkUserAccess(p.UserID); err != nil {
		return nil, err
	}

	msg := &dbmodels.UserMessage{
		Title:       p.Title,
		Contents:    p.Message,
		ToUser:      p.UserID,
		FromUserId:  0,
		FromSpaceId: c.spaceId,
		IsRead:      false,
	}

	coreHub := c.app.CoreHub().(*corehub.CoreHub)
	id, err := coreHub.UserSendMessage(msg)
	if err != nil {
		return nil, err
	}

	return map[string]any{"id": id}, nil
}
