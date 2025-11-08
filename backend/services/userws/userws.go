package userws

import "net"

type IUserWs interface {
	SendUser(userId int64, message string) error
	BroadcastGroup(groupName string, message string) error
	BroadcastAll(message string) error
	AddUserConnection(userId int64, groupName string, conn net.Conn) error
	RemoveUserConnection(userId int64, groupName string, conn net.Conn) error
}
