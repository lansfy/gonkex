package aerospike

import (
	"github.com/aerospike/aerospike-client-go/v5"
)

type Client struct {
	*aerospike.Client
	namespace string
}

func New(client *aerospike.Client, namespace string) *Client {
	return &Client{
		Client:    client,
		namespace: namespace,
	}
}

func (c *Client) Truncate(set string) error {
	return c.Client.Truncate(nil, c.namespace, set, nil)
}

func (c *Client) InsertBinMap(set, key string, binMap map[string]interface{}) error {
	aerospikeKey, err := aerospike.NewKey(c.namespace, set, key)
	if err != nil {
		return err
	}
	bins := prepareBins(binMap)

	return c.PutBins(nil, aerospikeKey, bins...)
}

func prepareBins(binmap map[string]interface{}) []*aerospike.Bin {
	var bins []*aerospike.Bin
	for binName, binData := range binmap {
		if binName == "$extend" {
			continue
		}
		bins = append(bins, aerospike.NewBin(binName, binData))
	}

	return bins
}
