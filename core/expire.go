package core

import (
	"log"
	"time"
)

func expireSample() float32 {

	var expiredCount int = 0
	var limit int = 20 // 20 random keys to be deleted

	for key, obj := range dataStore {
		if obj.expiryAt != -1 {
			limit--

			if obj.expiryAt <= time.Now().UnixMilli() {
				delete(dataStore, key)
				expiredCount++
			}

			if limit == 0 {
				break
			}

		}
	}

	return float32(expiredCount) / float32(limit)

}

func DeleteExpiredKeys() {
	for {
		frac := expireSample()

		if frac < 0.25 {
			break
		}

	}

	log.Println("deleted the expired but undeleted keys. total keys", len(dataStore))
}
