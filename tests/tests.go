package main

import "github.com/blue-monads/turnix/tests/stateless"

func main() {
	HandleUfsTest()
	HandleLuazUfsTest()
	HandleLuazMcpTest()
	stateless.RunStateLessLua()

}
