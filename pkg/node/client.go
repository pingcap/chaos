package node

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// Client is used to communicate with the node server
type Client struct {
	node   string
	addr   string
	client *http.Client
}

// NewClient creates the client.
func NewClient(nodeName string, addr string) *Client {
	return &Client{
		node:   nodeName,
		addr:   addr,
		client: http.DefaultClient,
	}
}

// Node returns the node name.
func (c *Client) Node() string {
	return c.node
}

func (c *Client) getURLPrefix() string {
	return fmt.Sprintf("http://%s/node", c.addr)
}

func (c *Client) doPost(suffix string, args url.Values, data []byte) error {
	if args == nil {
		args = url.Values{}
	}
	args.Set("node", c.node)
	url := fmt.Sprintf("%s%s?%s", c.getURLPrefix(), suffix, args.Encode())
	resp, err := c.client.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode%200 != 0 {
		return fmt.Errorf("%s:%s", resp.Status, data)
	}

	return nil
}

// SetUpDB is to set up the db
func (c *Client) SetUpDB(name string, nodes []string) error {
	v := url.Values{}
	v.Set("nodes", strings.Join(nodes, ","))

	return c.doPost(fmt.Sprintf("/db/%s/setup", name), v, nil)
}

// TearDownDB tears down db
func (c *Client) TearDownDB(name string, nodes []string) error {
	v := url.Values{}
	v.Set("nodes", strings.Join(nodes, ","))

	return c.doPost(fmt.Sprintf("/db/%s/teardown", name), v, nil)
}

// StartDB starts db
func (c *Client) StartDB(name string) error {
	return c.doPost(fmt.Sprintf("/db/%s/start", name), nil, nil)
}

// StopDB stops db
func (c *Client) StopDB(name string) error {
	return c.doPost(fmt.Sprintf("/db/%s/stop", name), nil, nil)
}

// KillDB kills db
func (c *Client) KillDB(name string) error {
	return c.doPost(fmt.Sprintf("/db/%s/kill", name), nil, nil)
}

// IsDBRunning checks db is running
func (c *Client) IsDBRunning(name string) bool {
	return c.doPost(fmt.Sprintf("/db/%s/is_running", name), nil, nil) == nil
}

// StartNemesis starts nemesis
func (c *Client) StartNemesis(name string, args ...string) error {
	v := url.Values{}
	suffix := fmt.Sprintf("/nemesis/%s/start", name)
	if len(args) > 0 {
		v.Set("args", strings.Join(args, ","))
	}
	return c.doPost(suffix, v, nil)
}

// StopNemesis stops nemesis
func (c *Client) StopNemesis(name string, args ...string) error {
	v := url.Values{}
	suffix := fmt.Sprintf("/nemesis/%s/stop", name)
	if len(args) > 0 {
		v.Set("args", strings.Join(args, ","))
	}
	return c.doPost(suffix, v, nil)
}
