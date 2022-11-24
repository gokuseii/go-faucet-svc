/*
 * GENERATED. Do not modify. Your changes might be overwritten!
 */

package resources

type SendNear struct {
	Key
	Attributes SendNearAttributes `json:"attributes"`
}
type SendNearResponse struct {
	Data     SendNear `json:"data"`
	Included Included `json:"included"`
}

type SendNearListResponse struct {
	Data     []SendNear `json:"data"`
	Included Included   `json:"included"`
	Links    *Links     `json:"links"`
}

// MustSendNear - returns SendNear from include collection.
// if entry with specified key does not exist - returns nil
// if entry with specified key exists but type or ID mismatches - panics
func (c *Included) MustSendNear(key Key) *SendNear {
	var sendNear SendNear
	if c.tryFindEntry(key, &sendNear) {
		return &sendNear
	}
	return nil
}
