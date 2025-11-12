package main

import "github.com/blue-monads/turnix/backend/utils/qq"

func main() {

	qq.Println("WebdavServer/start")

	server := NewWebdavServer(
		"localhost", 8666, "./tmp/buddyfs")
	server.Listen()

	qq.Println("WebdavServer/end")
}
