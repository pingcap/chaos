package cluster

import (
	"context"
	"fmt"
	"log"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/siddontang/chaos/pkg/util"
	"github.com/siddontang/chaos/pkg/util/ssh"
)

const (
	archiveURL = "http://download.pingcap.org/tidb-latest-linux-amd64.tar.gz"
	deployDir  = "/opt/tidb"
)

var (
	pdBinary   = path.Join(deployDir, "./bin/pd-server")
	tikvBinary = path.Join(deployDir, "./bin/tikv-server")
	tidbBinary = path.Join(deployDir, "./bin/tidb-server")

	pdConfig   = path.Join(deployDir, "./conf/pd.toml")
	tikvConfig = path.Join(deployDir, "./conf/tikv.toml")

	pdLog   = path.Join(deployDir, "./log/pd.log")
	tikvLog = path.Join(deployDir, "./log/tikv.log")
	tidbLog = path.Join(deployDir, "./log/tidb.log")
)

// Cluster is the TiKV/TiDB database cluster.
// Note: Cluster does not implement `core.DB` interface.
type Cluster struct {
	nodes          []string
	installBlocker util.BlockRunner
	IncludeTidb    bool
}

// SetUp initializes the database.
func (cluster *Cluster) SetUp(ctx context.Context, nodes []string, node string) error {
	// Try kill all old servers
	if cluster.IncludeTidb {
		ssh.Exec(ctx, node, "killall", "-9", "tidb-server")
	}
	ssh.Exec(ctx, node, "killall", "-9", "tikv-server")
	ssh.Exec(ctx, node, "killall", "-9", "pd-server")

	cluster.nodes = nodes

	cluster.installBlocker.Init(len(nodes))

	log.Printf("install archieve on node %s", node)

	var err error
	cluster.installBlocker.Run(func() {
		err = util.InstallArchive(ctx, node, archiveURL, deployDir)
	})
	if err != nil {
		return err
	}

	util.Mkdir(ctx, node, path.Join(deployDir, "conf"))
	util.Mkdir(ctx, node, path.Join(deployDir, "log"))

	pdCfs := []string{
		"tick-interval=\"100ms\"",
		"election-interval=\"500ms\"",
		"tso-save-interval=\"500ms\"",
		"[replication]",
		"max-replicas=5",
	}

	if err := util.WriteFile(ctx, node, pdConfig, strconv.Quote(strings.Join(pdCfs, "\n"))); err != nil {
		return err
	}

	tikvCfs := []string{
		"[raftstore]",
		"pd-heartbeat-tick-interval=\"500ms\"",
		"pd-store-heartbeat-tick-interval=\"1s\"",
		"raft_store_max_leader_lease=\"900ms\"",
		"raft_base_tick_interval=\"100ms\"",
		"raft_heartbeat_ticks=3",
		"raft_election_timeout_ticks=10",
	}

	if err := util.WriteFile(ctx, node, tikvConfig, strconv.Quote(strings.Join(tikvCfs, "\n"))); err != nil {
		return err
	}

	return cluster.start(ctx, node, true)
}

// TearDown tears down the database.
func (cluster *Cluster) TearDown(ctx context.Context, nodes []string, node string) error {
	return cluster.Kill(ctx, node)
}

// Start starts the database
func (cluster *Cluster) Start(ctx context.Context, node string) error {
	return cluster.start(ctx, node, false)
}

func (cluster *Cluster) start(ctx context.Context, node string, inSetUp bool) error {
	log.Printf("start database on node %s", node)

	initialClusterArgs := make([]string, len(cluster.nodes))
	for i, n := range cluster.nodes {
		initialClusterArgs[i] = fmt.Sprintf("%s=http://%s:2380", n, n)
	}
	pdArgs := []string{
		fmt.Sprintf("--name=%s", node),
		"--data-dir=pd",
		"--client-urls=http://0.0.0.0:2379",
		"--peer-urls=http://0.0.0.0:2380",
		fmt.Sprintf("--advertise-client-urls=http://%s:2379", node),
		fmt.Sprintf("--advertise-peer-urls=http://%s:2380", node),
		fmt.Sprintf("--initial-cluster=%s", strings.Join(initialClusterArgs, ",")),
		fmt.Sprintf("--log-file=%s", pdLog),
		fmt.Sprintf("--config=%s", pdConfig),
	}

	log.Printf("start pd-server on node %s", node)
	pdPID := path.Join(deployDir, "pd.pid")
	opts := util.NewDaemonOptions(deployDir, pdPID)
	if err := util.StartDaemon(ctx, node, opts, pdBinary, pdArgs...); err != nil {
		return err
	}

	if inSetUp {
		time.Sleep(5 * time.Second)
	}

	if !util.IsDaemonRunning(ctx, node, pdBinary, pdPID) {
		return fmt.Errorf("fail to start pd on node %s", node)
	}

	// Before starting TiKV, we should ensure PD cluster is ready.
	for {
		// `--fail`, non-zero exit code on server errors.
		_, err := ssh.CombinedOutput(ctx, node, "curl", "--fail localhost:2379/pd/api/v1/members")
		if err == nil {
			break
		}
		log.Println("wait pd cluster...")
		time.Sleep(1 * time.Second)
	}

	pdEndpoints := make([]string, len(cluster.nodes))
	for i, n := range cluster.nodes {
		pdEndpoints[i] = fmt.Sprintf("%s:2379", n)
	}

	tikvArgs := []string{
		fmt.Sprintf("--pd=%s", strings.Join(pdEndpoints, ",")),
		"--addr=0.0.0.0:20160",
		fmt.Sprintf("--advertise-addr=%s:20160", node),
		"--data-dir=tikv",
		fmt.Sprintf("--log-file=%s", tikvLog),
		fmt.Sprintf("--config=%s", tikvConfig),
	}

	log.Printf("start tikv-server on node %s", node)
	tikvPID := path.Join(deployDir, "tikv.pid")
	opts = util.NewDaemonOptions(deployDir, tikvPID)
	if err := util.StartDaemon(ctx, node, opts, tikvBinary, tikvArgs...); err != nil {
		return err
	}

	if inSetUp {
		time.Sleep(30 * time.Second)
	}

	if !util.IsDaemonRunning(ctx, node, tikvBinary, tikvPID) {
		return fmt.Errorf("fail to start tikv on node %s", node)
	}

	if cluster.IncludeTidb {
		tidbArgs := []string{
			"--store=tikv",
			fmt.Sprintf("--path=%s", strings.Join(pdEndpoints, ",")),
			fmt.Sprintf("--log-file=%s", tidbLog),
		}

		log.Printf("start tidb-server on node %s", node)
		tidbPID := path.Join(deployDir, "tidb.pid")
		opts = util.NewDaemonOptions(deployDir, tidbPID)
		if err := util.StartDaemon(ctx, node, opts, tidbBinary, tidbArgs...); err != nil {
			return err
		}

		var err error
		if inSetUp {
			for i := 0; i < 12; i++ {
				if err = ssh.Exec(ctx, node, "curl", fmt.Sprintf("http://%s:10080/status", node)); err == nil {
					break
				}
				log.Printf("try to wait tidb run on %s", node)
				time.Sleep(10 * time.Second)
			}
		}

		if err != nil {
			return err
		}

		if !util.IsDaemonRunning(ctx, node, tidbBinary, tidbPID) {
			return fmt.Errorf("fail to start tidb on node %s", node)
		}
	}

	return nil
}

// Stop stops the database
func (cluster *Cluster) Stop(ctx context.Context, node string) error {
	if cluster.IncludeTidb {
		if err := util.StopDaemon(ctx, node, tidbBinary, path.Join(deployDir, "tidb.pid")); err != nil {
			return err
		}
	}

	if err := util.StopDaemon(ctx, node, tikvBinary, path.Join(deployDir, "tikv.pid")); err != nil {
		return err
	}

	return util.StopDaemon(ctx, node, pdBinary, path.Join(deployDir, "pd.pid"))
}

// Kill kills the database
func (cluster *Cluster) Kill(ctx context.Context, node string) error {
	if cluster.IncludeTidb {
		if err := util.KillDaemon(ctx, node, tidbBinary, path.Join(deployDir, "tidb.pid")); err != nil {
			return err
		}
	}

	if err := util.KillDaemon(ctx, node, tikvBinary, path.Join(deployDir, "tikv.pid")); err != nil {
		return err
	}

	return util.KillDaemon(ctx, node, pdBinary, path.Join(deployDir, "pd.pid"))
}

// IsRunning checks whether the database is running or not
func (cluster *Cluster) IsRunning(ctx context.Context, node string) bool {
	if cluster.IncludeTidb {
		return util.IsDaemonRunning(ctx, node, tidbBinary, path.Join(deployDir, "tidb.pid"))
	}
	return util.IsDaemonRunning(ctx, node, tidbBinary, path.Join(deployDir, "tikv.pid"))
}

// Name returns the unique name for the database
func (cluster *Cluster) Name() string {
	return "cluster"
}
