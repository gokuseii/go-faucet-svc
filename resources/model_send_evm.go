/*
 * GENERATED. Do not modify. Your changes might be overwritten!
 */

package resources

type SendEvm struct {
	Key
	Attributes SendEvmAttributes `json:"attributes"`
}
type SendEvmResponse struct {
	Data     SendEvm  `json:"data"`
	Included Included `json:"included"`
}

type SendEvmListResponse struct {
	Data     []SendEvm `json:"data"`
	Included Included  `json:"included"`
	Links    *Links    `json:"links"`
}

// MustSendEvm - returns SendEvm from include collection.
// if entry with specified key does not exist - returns nil
// if entry with specified key exists but type or ID mismatches - panics
func (c *Included) MustSendEvm(key Key) *SendEvm {
	var sendEvm SendEvm
	if c.tryFindEntry(key, &sendEvm) {
		return &sendEvm
	}
	return nil
}
