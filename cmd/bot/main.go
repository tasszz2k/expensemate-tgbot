package main

import (
	"expensemate-tgbot/internal/servers"
	"expensemate-tgbot/pkg/configs"
)

func main() {
	expensemateBot := servers.NewServer(servers.ServerConfig{AppConf: configs.Get()})
	// start `expensemate bot`
	if err := expensemateBot.Start(); err != nil {
		panic(err)
	}
}
