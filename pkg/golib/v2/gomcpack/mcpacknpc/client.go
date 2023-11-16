package mcpacknpc

import (
	"bytes"

	"permission/pkg/golib/v2/gomcpack/mcpack"
	"permission/pkg/golib/v2/gomcpack/npc"
)

type Client struct {
	*npc.Client
}

func NewClient(server []string) *Client {
	c := npc.NewClient(server)
	return &Client{Client: c}
}

func (c *Client) Call(args interface{}, reply interface{}) error {
	content, err := mcpack.Marshal(args)
	if err != nil {
		return err
	}
	resp, err := c.Client.Do(npc.NewRequest(bytes.NewReader(content)))
	if err != nil {
		return err
	}
	return mcpack.Unmarshal(resp.Body, reply)
}

func (c *Client) Send(args []byte) ([]byte, error) {
	resp, err := c.Client.Do(npc.NewRequest(bytes.NewReader(args)))
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
