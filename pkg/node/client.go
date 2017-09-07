package node

import "net/http"
import "fmt"
import "io/ioutil"
import "bytes"

import "strings"

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
		client: &http.Client{},
	}
}

func (c *Client) getURLPrefix() string {
	return fmt.Sprintf("http://%s/node", c.addr)
}

func (c *Client) doPost(suffix string, data []byte) error {
	url := fmt.Sprintf("%s%s", c.getURLPrefix(), suffix)
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
func (c *Client) SetUpDB(name string) error {
	return c.doPost(fmt.Sprintf("/db/%s/setup", name), nil)
}

// TearDownDB tears down db
func (c *Client) TearDownDB(name string) error {
	return c.doPost(fmt.Sprintf("/db/%s/teardown", name), nil)
}

// StartDB starts db
func (c *Client) StartDB(name string) error {
	return c.doPost(fmt.Sprintf("/db/%s/start", name), nil)
}

// StopDB stops db
func (c *Client) StopDB(name string) error {
	return c.doPost(fmt.Sprintf("/db/%s/stop", name), nil)
}

// KillDB kills db
func (c *Client) KillDB(name string) error {
	return c.doPost(fmt.Sprintf("/db/%s/kill", name), nil)
}

// IsDBRunning checks db is running
func (c *Client) IsDBRunning(name string) bool {
	return c.doPost(fmt.Sprintf("/db/%s/is_running", name), nil) == nil
}

// // SetUpNemesis is to set up the nemesis
// func (c *Client) SetUpNemesis(name string) error {
// 	return c.doPost(fmt.Sprintf("/nemesis/%s/setup", name), nil)
// }

// // TearDownNemesis tears down nemesis
// func (c *Client) TearDownNemesis(name string) error {
// 	return c.doPost(fmt.Sprintf("/nemesis/%s/teardown", name), nil)
// }

// StartNemesis starts nemesis
func (c *Client) StartNemesis(name string, args ...string) error {
	suffix := fmt.Sprintf("/nemesis/%s/start", name)
	if len(args) > 0 {
		suffix = fmt.Sprintf("%s?args=%s", suffix, strings.Join(args, ","))
	}
	return c.doPost(suffix, nil)
}

// StopNemesis stops nemesis
func (c *Client) StopNemesis(name string, args ...string) error {
	suffix := fmt.Sprintf("/nemesis/%s/stop", name)
	if len(args) > 0 {
		suffix = fmt.Sprintf("%s?args=%s", suffix, strings.Join(args, ","))
	}
	return c.doPost(suffix, nil)
}