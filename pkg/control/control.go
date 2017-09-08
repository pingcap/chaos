package control

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/siddontang/chaos/pkg/core"
	"github.com/siddontang/chaos/pkg/node"
)

// Controller controls the whole cluster. It sends request to the database,
// and also uses nemesis to disturb the cluster.
// Here have only 5 nodes, and the hosts are n1 - n5.
type Controller struct {
	db string

	nodes       []string
	nodeClients []*node.Client

	clients          []core.Client
	requestGenerator []core.RequestGenerator

	nemesisGenerators []core.NemesisGenerator

	ctx    context.Context
	cancel context.CancelFunc
}

// NewController creates a controller.
// nodePort is used to communicate with the node server.
// db is the name which we want to run, you must register the db in the node before.
// clientCreator creates the client to communicate with the db.
func NewController(nodePort int, db string, clientCreator core.ClientCreator) *Controller {
	c := new(Controller)
	c.db = db
	c.ctx, c.cancel = context.WithCancel(context.Background())

	for i := 1; i <= 5; i++ {
		name := fmt.Sprintf("n%d", i)
		c.nodes = append(c.nodes, name)
		client := node.NewClient(name, fmt.Sprintf("%s:%d", name, nodePort))
		c.nodeClients = append(c.nodeClients, client)
		c.clients = append(c.clients, clientCreator.Create(name))
		c.requestGenerator = append(c.requestGenerator, clientCreator.CreateRequestGenerator())
	}

	return c
}

// Close closes the controller.
func (c *Controller) Close() {
	c.cancel()
}

// AddNemesisGenerator adds a nemesis generator.
func (c *Controller) AddNemesisGenerator(g core.NemesisGenerator) {
	c.nemesisGenerators = append(c.nemesisGenerators, g)
}

// Run runs the controller.
// requestCount controls how many requests a client sends to the db and runTime controls how
// long the controller takes. If any exceeds the value, we will stop the controller.
func (c *Controller) Run(requestCount int, runTime time.Duration) {
	if requestCount == 0 {
		requestCount = 10000
	}

	if runTime == 0 {
		runTime = time.Minute * 5
	}

	if len(c.requestGenerator) == 0 {
		log.Fatal("must add a request generator")
	}

	c.setUpDB()

	n := len(c.nodes)
	var clientWg sync.WaitGroup
	clientWg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer clientWg.Done()
			c.onClientLoop(i, requestCount, runTime)
		}(i)
	}

	ctx, cancel := context.WithCancel(c.ctx)
	var nemesisWg sync.WaitGroup
	nemesisWg.Add(1)
	go func() {
		defer nemesisWg.Done()
		c.dispatchNemesis(ctx)
	}()

	clientWg.Wait()

	cancel()
	nemesisWg.Wait()

	c.tearDownDB()
}

func (c *Controller) setUpDB() {
	log.Printf("begin to set up database")
	var wg sync.WaitGroup
	n := len(c.nodes)
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			client := c.nodeClients[i]
			log.Printf("begin to set up database on %s", c.nodes[i])
			if err := client.SetUpDB(c.db); err != nil {
				log.Fatalf("setup db %s at node %s failed %v", c.db, c.nodes[i], err)
			}
		}(i)
	}
	wg.Wait()
}

func (c *Controller) tearDownDB() {
	log.Printf("begin to tear down database")
	var wg sync.WaitGroup
	n := len(c.nodes)
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			client := c.nodeClients[i]
			log.Printf("being to tear down database on %s", c.nodes[i])
			if err := client.TearDownDB(c.db); err != nil {
				log.Fatalf("tear down db %s at node %s failed %v", c.db, c.nodes[i], err)
			}
		}(i)
	}
	wg.Wait()
}

func (c *Controller) onClientLoop(i int, requestCount int, runTime time.Duration) {
	client := c.clients[i]
	gen := c.requestGenerator[i]
	node := c.nodes[i]

	ctx, cancel := context.WithTimeout(c.ctx, runTime)
	defer cancel()

	log.Printf("begin to set up db client for node %s", node)
	if err := client.SetUp(ctx, node); err != nil {
		log.Printf("set up cdb lient for node %s failed %v", node, err)
	}

	defer func() {
		log.Printf("begin to tear down db client for node %s", node)
		if err := client.TearDown(c.ctx, node); err != nil {
			log.Printf("tear down db client for node %s failed %v", node, err)
		}
	}()

	for i := 0; i < requestCount; i++ {
		request := gen.Generate()

		// TODO: add to history
		_, err := client.Invoke(ctx, node, request)
		if err != nil {
			log.Printf("invoke db client for node %s failed %v", node, err)
		}
	}
}

func (c *Controller) dispatchNemesis(ctx context.Context) {
	if len(c.nemesisGenerators) == 0 {
		return
	}

	log.Printf("begin to run nemesis")
	var wg sync.WaitGroup
	n := len(c.nodes)
LOOP:
	for {
		for _, g := range c.nemesisGenerators {
			select {
			case <-ctx.Done():
				break LOOP
			default:
			}

			log.Printf("begin to run %s nemesis generator", g)
			ops := g.Generate()

			wg.Add(n)
			for i := 0; i < n; i++ {
				go c.onNemesisLoop(ctx, i, ops[i], &wg)
			}
			wg.Wait()
		}
	}
	log.Printf("stop to run nemesis")
}

func (c *Controller) onNemesisLoop(ctx context.Context, index int, op core.NemesisOperation, wg *sync.WaitGroup) {
	defer wg.Done()

	nodeClient := c.nodeClients[index]
	node := c.nodes[index]

	log.Printf("start nemesis %s with %v on %s", op.Name, op.StartArgs, node)
	if err := nodeClient.StartNemesis(op.Name, op.StartArgs...); err != nil {
		log.Printf("start nemesis %s with %v on %s failed: %v", op.Name, op.StartArgs, node, err)
	}

	select {
	case <-time.After(op.WaitTime):
	case <-ctx.Done():
	}

	log.Printf("stop nemesis %s with %v on %s", op.Name, op.StartArgs, node)
	if err := nodeClient.StopNemesis(op.Name, op.StopArgs...); err != nil {
		log.Printf("stop nemesis %s with %v on %s failed: %v", op.Name, op.StopArgs, node, err)
	}
}
