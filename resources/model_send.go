/*
 * GENERATED. Do not modify. Your changes might be overwritten!
 */

package resources

type Send struct {
	Key
	Attributes SendAttributes `json:"attributes"`
}
type SendResponse struct {
	Data     Send     `json:"data"`
	Included Included `json:"included"`
}

type SendListResponse struct {
	Data     []Send   `json:"data"`
	Included Included `json:"included"`
	Links    *Links   `json:"links"`
}

// MustSend - returns Send from include collection.
// if entry with specified key does not exist - returns nil
// if entry with specified key exists but type or ID mismatches - panics
func (c *Included) MustSend(key Key) *Send {
	var send Send
	if c.tryFindEntry(key, &send) {
		return &send
	}
	return nil
}
