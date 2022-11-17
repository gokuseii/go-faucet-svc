/*
 * GENERATED. Do not modify. Your changes might be overwritten!
 */

package resources

type SendSolana struct {
	Key
	Attributes SendSolanaAttributes `json:"attributes"`
}
type SendSolanaResponse struct {
	Data     SendSolana `json:"data"`
	Included Included   `json:"included"`
}

type SendSolanaListResponse struct {
	Data     []SendSolana `json:"data"`
	Included Included     `json:"included"`
	Links    *Links       `json:"links"`
}

// MustSendSolana - returns SendSolana from include collection.
// if entry with specified key does not exist - returns nil
// if entry with specified key exists but type or ID mismatches - panics
func (c *Included) MustSendSolana(key Key) *SendSolana {
	var sendSolana SendSolana
	if c.tryFindEntry(key, &sendSolana) {
		return &sendSolana
	}
	return nil
}
