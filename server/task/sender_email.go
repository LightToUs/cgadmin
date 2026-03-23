package task

import "github.com/flipped-aurora/gin-vue-admin/server/service/system"

func SenderEmailDailyReset() error {
	return (&system.SenderEmailAccountService{}).DailyResetAll()
}

func SenderEmailConnectivityCheck() error {
	return (&system.SenderEmailAccountService{}).ConnectivityCheckAll()
}
