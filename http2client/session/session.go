package session

import "sync"

var (
	localIds = struct {
		sync.RWMutex
		m map[uint32]bool
	}{m: make(map[uint32]bool)}
)

func init() {
	// StreamId 0 is reserved
	SetLocalSteamId(0)
}

func SetLocalSteamId(id uint32) {
	localIds.Lock()
	localIds.m[id] = true
	localIds.Unlock()
}

func IsLocalSteamId(id uint32) bool {

	localIds.RLock()
	_, exists := localIds.m[id]
	localIds.RUnlock()
	return exists
}

func ResetSession() {
	// TODO: Reset id's if the TCP connection goes down

	for key, _ := range localIds.m {
		delete(localIds.m, key)
	}

}
