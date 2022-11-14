/*
 * GENERATED. Do not modify. Your changes might be overwritten!
 */

package resources

type SendEvmToken struct {
	Key
	Attributes SendEvmTokenAttributes `json:"attributes"`
}
type SendEvmTokenResponse struct {
	Data     SendEvmToken `json:"data"`
	Included Included     `json:"included"`
}

type SendEvmTokenListResponse struct {
	Data     []SendEvmToken `json:"data"`
	Included Included       `json:"included"`
	Links    *Links         `json:"links"`
}

// MustSendEvmToken - returns SendEvmToken from include collection.
// if entry with specified key does not exist - returns nil
// if entry with specified key exists but type or ID mismatches - panics
func (c *Included) MustSendEvmToken(key Key) *SendEvmToken {
	var sendEvmToken SendEvmToken
	if c.tryFindEntry(key, &sendEvmToken) {
		return &sendEvmToken
	}
	return nil
}
