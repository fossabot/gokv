package hazelcast

import (
	"time"

	hazelcast "github.com/hazelcast/hazelcast-go-client"

	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/util"
)

var defaultTimeout = 200 * time.Millisecond

// Client is a gokv.Store implementation for Hazelcast.
type Client struct {
	c     *hazelcast.Client
	codec encoding.Codec
}

// Set stores the given value for the given key.
// Values are automatically marshalled to JSON or gob (depending on the configuration).
// The key must not be "" and the value must not be nil.
func (c Client) Set(k string, v interface{}) error {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return err
	}

	// First turn the passed object into something that Hazelcast can handle
	data, err := c.codec.Marshal(v)
	if err != nil {
		return err
	}

	item := memcache.Item{
		Key:   k,
		Value: data,
	}
	err = c.c.Set(&item)
	if err != nil {
		return err
	}

	return nil
}

// Get retrieves the stored value for the given key.
// You need to pass a pointer to the value, so in case of a struct
// the automatic unmarshalling can populate the fields of the object
// that v points to with the values of the retrieved object's values.
// If no value is found it returns (false, nil).
// The key must not be "" and the pointer must not be nil.
func (c Client) Get(k string, v interface{}) (found bool, err error) {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return false, err
	}

	item, err := c.c.Get(k)
	// If no value was found return false
	if err == memcache.ErrCacheMiss {
		return false, nil
	} else if err != nil {
		return false, err
	}
	data := item.Value

	return true, c.codec.Unmarshal(data, v)
}

// Delete deletes the stored value for the given key.
// The key must not be longer than 250 bytes (this is a restriction of Hazelcast).
// Deleting a non-existing key-value pair does NOT lead to an error.
// The key must not be "".
func (c Client) Delete(k string) error {
	if err := util.CheckKey(k); err != nil {
		return err
	}

	err := c.c.Delete(k)
	if err == memcache.ErrCacheMiss {
		return nil
	}
	return err
}

// Close closes the client.
// In the Hazelcast implementation this doesn't have any effect.
func (c Client) Close() error {
	return nil
}

// Options are the options for the Hazelcast client.
type Options struct {
	// Address of one Hazelcast server, including port.
	// Optional ("localhost:5701" by default).
	Address string
	// Encoding format.
	// Optional (encoding.JSON by default).
	Codec encoding.Codec
}

// DefaultOptions is an Options object with default values.
// Addresses: "localhost:11211", Timeout: 200 milliseconds, MaxIdleConns: 100, Codec: encoding.JSON
var DefaultOptions = Options{
	Address:    "localhost:5701",
	Codec:        encoding.JSON,
}

// NewClient creates a new Hazelcast client.
func NewClient(options Options) (Client, error) {
	result := Client{}

	// Set default values
	if options.Address == nil{
		options.Address = DefaultOptions.Address
	}
	if options.Codec == nil {
		options.Codec = DefaultOptions.Codec
	}

	config := hazelcast.NewConfig()
	config.NetworkConfig().AddAddress(options.Address)
	client , err := hazelcast.NewClientWithConfig(config)
	if err != nil{
		return result, err
	}

	result.c = client
	result.codec = options.Codec

	return result, nil
}
