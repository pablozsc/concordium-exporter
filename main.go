package main

import (
	"context"
	"encoding/json"
	"reflect"

	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	pb "github.com/liray-unendlich/concordium-grpc-api"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// Setup initial variables
var (
	url      string
	password string
	bakerYn  bool
)

var version = "1.3.0"

const namespace = "concordium"

type metricsData struct {
	PeerTotalReceived        float64
	PeerTotalSent            float64
	BestBlock                string
	LastFinalizedBlockHeight float64
	BlockArriveLatencyEMSD   float64
	BlockReceiveLatencyEMSD  float64
	BlockReceivePeriodEMSD   float64
	BlockArrivePeriodEMSD    float64
	BlocksReceivedCount      float64
	TransactionsPerBlockEMSD float64
	FinalizationPeriodEMA    float64
	BestBlockHeight          float64
	FinalizationCount        float64
	EpochDuration            float64
	BlocksVerifiedCount      float64
	SlotDuration             float64
	FinalizationPeriodEMSD   float64
	TransactionsPerBlockEMA  float64
	BlockArriveLatencyEMA    float64
	BlockReceiveLatencyEMA   float64
	BlockArrivePeriodEMA     float64
	BlockReceivePeriodEMA    float64
	BakerRunning             int
	Running                  int
	BakerId                  int
	BakerLotteryPower        float64
	EstimatedBakingBlock     float64
}

type concordiumCollector struct {
	PeerTotalSent            prometheus.Gauge
	PeerTotalReceived        prometheus.Gauge
	LastFinalizedBlockHeight prometheus.Gauge
	BlockArriveLatencyEMSD   prometheus.Gauge
	BlockReceiveLatencyEMSD  prometheus.Gauge
	BlockReceivePeriodEMSD   prometheus.Gauge
	BlockArrivePeriodEMSD    prometheus.Gauge
	BlocksReceivedCount      prometheus.Gauge
	TransactionsPerBlockEMSD prometheus.Gauge
	FinalizationPeriodEMA    prometheus.Gauge
	BestBlockHeight          prometheus.Gauge
	LastFinalizedTime        prometheus.Gauge
	FinalizationCount        prometheus.Gauge
	EpochDuration            prometheus.Gauge
	BlocksVerifiedCount      prometheus.Gauge
	SlotDuration             prometheus.Gauge
	FinalizationPeriodEMSD   prometheus.Gauge
	TransactionsPerBlockEMA  prometheus.Gauge
	BlockArriveLatencyEMA    prometheus.Gauge
	BlockReceiveLatencyEMA   prometheus.Gauge
	BlockArrivePeriodEMA     prometheus.Gauge
	BlockReceivePeriodEMA    prometheus.Gauge
	BakerRunning             prometheus.Gauge
	Running                  prometheus.Gauge
	BakerId                  prometheus.Gauge
	BakerLotteryPower        prometheus.Gauge
	EstimatedBakingBlock     prometheus.Gauge
}

func newconcordiumCollector() *concordiumCollector {
	return &concordiumCollector{
		PeerTotalSent: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "peer_total_sent_amount",
			Help:      "Peer total sent packets in Byte",
		}),
		PeerTotalReceived: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "peer_total_received_amount",
			Help:      "Peer total received packets in Byte",
		}),
		LastFinalizedBlockHeight: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "last_finalized_block_height",
			Help:      "Last finalized block height",
		}),
		BlockArriveLatencyEMSD: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "block_arrive_latency_inEMSD",
			Help:      "Arrived block latency in EMSD",
		}),
		BlockReceiveLatencyEMSD: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "block_receive_latency_inEMSD",
			Help:      "Received block latency in EMSD",
		}),
		BlockReceivePeriodEMSD: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "block_receive_period_inEMSD",
			Help:      "Received block period in EMSD",
		}),
		BlockArrivePeriodEMSD: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "block_arrive_period_inEMSD",
			Help:      "Arrived block period in EMSD",
		}),
		BlocksReceivedCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "block_received_count",
			Help:      "Received block count",
		}),
		TransactionsPerBlockEMSD: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "transactions_per_block_inEMSD",
			Help:      "Transaction count per block in EMSD",
		}),
		FinalizationPeriodEMA: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "finalization_period_inEMA",
			Help:      "Finalization period in EMA",
		}),
		BestBlockHeight: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "best_block_height",
			Help:      "Best block height",
		}),
		FinalizationCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "finalization_count",
			Help:      "Finalization count",
		}),
		EpochDuration: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "epoch_duration",
			Help:      "Epoch duration(const.)",
		}),
		BlocksVerifiedCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "blocks_verified_count",
			Help:      "Verified blocks count",
		}),
		SlotDuration: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "slot_duration",
			Help:      "Slot duration(const.)",
		}),
		FinalizationPeriodEMSD: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "finalization_period_inEMSD",
			Help:      "Finalization period in EMSD",
		}),
		TransactionsPerBlockEMA: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "transactions_per_block_inEMA",
			Help:      "Transactions per block in EMA",
		}),
		BlockArriveLatencyEMA: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "block_arrive_latency_inEMA",
			Help:      "Arrived block latency in EMA",
		}),
		BlockReceiveLatencyEMA: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "block_receive_latency_inEMA",
			Help:      "Received block latency in EMA",
		}),
		BlockArrivePeriodEMA: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "block_arrive_period_inEMA",
			Help:      "Arrived block period in EMA",
		}),
		BlockReceivePeriodEMA: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "block_receive_period_inEMA",
			Help:      "Received block period in EMA",
		}),
		BakerRunning: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "baker_running",
			Help:      "Bool value of whether baker is running. true=1, false=0",
		}),
		Running: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "consensus_running",
			Help:      "Bool value of whether consensus module is running. true=1, false=0",
		}),
		BakerId: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "baker_id",
			Help:      "Baker ID in integer",
		}),
		BakerLotteryPower: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "baker_lottery_power",
			Help:      "Baker Block Minting Probability",
		}),
		EstimatedBakingBlock: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "estimated_baking_block_per_day",
			Help:      "The number of blocks your baker is expected to bake per day",
		}),
	}
}

func (c *concordiumCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.PeerTotalSent.Desc()
	ch <- c.PeerTotalReceived.Desc()
	ch <- c.LastFinalizedBlockHeight.Desc()
	ch <- c.BlockArriveLatencyEMSD.Desc()
	ch <- c.BlockReceiveLatencyEMSD.Desc()
	ch <- c.BlockReceivePeriodEMSD.Desc()
	ch <- c.BlockArrivePeriodEMSD.Desc()
	ch <- c.BlocksReceivedCount.Desc()
	ch <- c.TransactionsPerBlockEMSD.Desc()
	ch <- c.FinalizationPeriodEMA.Desc()
	ch <- c.BestBlockHeight.Desc()
	ch <- c.FinalizationCount.Desc()
	ch <- c.EpochDuration.Desc()
	ch <- c.BlocksVerifiedCount.Desc()
	ch <- c.SlotDuration.Desc()
	ch <- c.FinalizationPeriodEMSD.Desc()
	ch <- c.TransactionsPerBlockEMA.Desc()
	ch <- c.BlockArriveLatencyEMA.Desc()
	ch <- c.BlockReceiveLatencyEMA.Desc()
	ch <- c.BlockArrivePeriodEMA.Desc()
	ch <- c.BlockReceivePeriodEMA.Desc()
	ch <- c.BakerRunning.Desc()
	ch <- c.Running.Desc()
	ch <- c.BakerId.Desc()
	ch <- c.BakerLotteryPower.Desc()
	ch <- c.EstimatedBakingBlock.Desc()
}

func (c *concordiumCollector) Collect(ch chan<- prometheus.Metric) {
	data, err := c.fetchAPI()
	if err != nil {
		log.Printf("%v", err)
		return
	}

	log.Printf("Finished fetching metrics")
	ch <- prometheus.MustNewConstMetric(
		c.PeerTotalSent.Desc(),
		prometheus.GaugeValue,
		data.PeerTotalSent,
	)
	ch <- prometheus.MustNewConstMetric(
		c.PeerTotalReceived.Desc(),
		prometheus.GaugeValue,
		data.PeerTotalReceived,
	)
	ch <- prometheus.MustNewConstMetric(
		c.LastFinalizedBlockHeight.Desc(),
		prometheus.GaugeValue,
		data.LastFinalizedBlockHeight,
	)
	ch <- prometheus.MustNewConstMetric(
		c.BlockArriveLatencyEMSD.Desc(),
		prometheus.GaugeValue,
		data.BlockArriveLatencyEMSD,
	)
	ch <- prometheus.MustNewConstMetric(
		c.BlockReceiveLatencyEMSD.Desc(),
		prometheus.GaugeValue,
		data.BlockReceiveLatencyEMSD,
	)
	ch <- prometheus.MustNewConstMetric(
		c.BlockReceivePeriodEMSD.Desc(),
		prometheus.GaugeValue,
		data.BlockReceivePeriodEMSD,
	)
	ch <- prometheus.MustNewConstMetric(
		c.BlockArrivePeriodEMSD.Desc(),
		prometheus.GaugeValue,
		data.BlockArrivePeriodEMSD,
	)
	ch <- prometheus.MustNewConstMetric(
		c.BlocksReceivedCount.Desc(),
		prometheus.GaugeValue,
		data.BlocksReceivedCount,
	)
	ch <- prometheus.MustNewConstMetric(
		c.TransactionsPerBlockEMSD.Desc(),
		prometheus.GaugeValue,
		data.TransactionsPerBlockEMSD,
	)
	ch <- prometheus.MustNewConstMetric(
		c.FinalizationPeriodEMA.Desc(),
		prometheus.GaugeValue,
		data.FinalizationPeriodEMA,
	)
	ch <- prometheus.MustNewConstMetric(
		c.BestBlockHeight.Desc(),
		prometheus.GaugeValue,
		data.BestBlockHeight,
	)
	ch <- prometheus.MustNewConstMetric(
		c.FinalizationCount.Desc(),
		prometheus.GaugeValue,
		data.FinalizationCount,
	)
	ch <- prometheus.MustNewConstMetric(
		c.EpochDuration.Desc(),
		prometheus.GaugeValue,
		data.EpochDuration,
	)
	ch <- prometheus.MustNewConstMetric(
		c.SlotDuration.Desc(),
		prometheus.GaugeValue,
		data.SlotDuration,
	)
	ch <- prometheus.MustNewConstMetric(
		c.FinalizationPeriodEMSD.Desc(),
		prometheus.GaugeValue,
		data.FinalizationPeriodEMSD,
	)
	ch <- prometheus.MustNewConstMetric(
		c.TransactionsPerBlockEMA.Desc(),
		prometheus.GaugeValue,
		data.TransactionsPerBlockEMA,
	)
	ch <- prometheus.MustNewConstMetric(
		c.BlockArriveLatencyEMA.Desc(),
		prometheus.GaugeValue,
		data.BlockArriveLatencyEMA,
	)
	ch <- prometheus.MustNewConstMetric(
		c.BlockReceiveLatencyEMA.Desc(),
		prometheus.GaugeValue,
		data.BlockReceiveLatencyEMA,
	)
	ch <- prometheus.MustNewConstMetric(
		c.BlockArrivePeriodEMA.Desc(),
		prometheus.GaugeValue,
		data.BlockArrivePeriodEMA,
	)
	ch <- prometheus.MustNewConstMetric(
		c.BlockReceivePeriodEMA.Desc(),
		prometheus.GaugeValue,
		data.BlockReceivePeriodEMA,
	)
	ch <- prometheus.MustNewConstMetric(
		c.BakerRunning.Desc(),
		prometheus.GaugeValue,
		float64(data.BakerRunning),
	)
	ch <- prometheus.MustNewConstMetric(
		c.Running.Desc(),
		prometheus.GaugeValue,
		float64(data.Running),
	)
	ch <- prometheus.MustNewConstMetric(
		c.BakerId.Desc(),
		prometheus.GaugeValue,
		float64(data.BakerId),
	)
	ch <- prometheus.MustNewConstMetric(
		c.BakerLotteryPower.Desc(),
		prometheus.GaugeValue,
		float64(data.BakerLotteryPower),
	)
	ch <- prometheus.MustNewConstMetric(
		c.EstimatedBakingBlock.Desc(),
		prometheus.GaugeValue,
		float64(data.EstimatedBakingBlock),
	)
}

// fetchAPI fetches the results of GetConsensusStatus, PeerTotalSent, PeerTotalReceived
func (c *concordiumCollector) fetchAPI() (metricsData, error) {
	var data metricsData
	data.BakerRunning = 0
	data.Running = 0
	var bestblock pb.BlockHash

	type baker struct {
		BakerID           int     `json:"bakerId"`
		BakerLotteryPower float64 `json:"bakerLotteryPower"`
		BakerAccount      string  `json:"bakerAccount"`
	}

	type birkParam struct {
		ElectionDifficulty float64 `json:"electionDifficulty"`
		ElectionNonce      string  `json:"electionNonce"`
		Bakers             []baker
	}

	var bakerStatus birkParam

	log.Printf("Start fetching metrics")
	conn, err := grpc.Dial(url, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("client connection error: %#v", err)
	}
	defer conn.Close()
	md := metadata.New(map[string]string{"authentication": password})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	client := pb.NewP2PClient(conn)

	log.Printf("Calling GetConsensusStatus")
	// Call GetConsensusStatus
	response_con, err := client.GetConsensusStatus(ctx, &pb.Empty{})
	if err != nil {
		fmt.Printf("Get ConsensusStatus Error:%#v \n", err)
		return data, err
	}
	err = json.Unmarshal([]byte(response_con.Value), &data)

	bestblock.BlockHash = data.BestBlock

	log.Printf("Calling PeerTotalSent")
	// Call PeerTotalSent
	response_sent, err_sent := client.PeerTotalSent(ctx, &pb.Empty{})
	if err_sent != nil {
		fmt.Printf("Get PeerTotalSent Error:%#v \n", err_sent)
		return data, err_sent
	}
	data.PeerTotalSent = float64(response_sent.Value)

	log.Printf("Calling PeerTotalReceived")
	// Call PeerTotalReceived
	response_rec, err_rec := client.PeerTotalReceived(ctx, &pb.Empty{})
	if err_rec != nil {
		fmt.Printf("Get PeerTotalReceived Error:%#v \n", err_rec)
		return data, err_rec
	}
	data.PeerTotalReceived = float64(response_rec.Value)

	// Call NodeInfo
	response_node, err_node := client.NodeInfo(ctx, &pb.Empty{})
	if err_node != nil {
		fmt.Printf("Get NodeInfo:%#v \n", err_node)
		return data, err_node
	}
	// Judge baker status
	if response_node.ConsensusBakerRunning {
		data.BakerRunning = 1
	} else {
		data.BakerRunning = 0
	}
	// Judge node status
	if response_node.ConsensusRunning {
		data.Running = 1
	} else {
		data.Running = 0
	}
	// Exchange bakerId
	if bakerYn {
		if (response_node.ConsensusBakerId == nil) || reflect.ValueOf(response_node.ConsensusBakerId).IsNil() {
			data.BakerId = 0
		} else {
			data.BakerId = int(response_node.ConsensusBakerId.Value)
			log.Printf("Calling BirkParameters")
			// Call BirkInfo
			response_birk, err_birk := client.GetBirkParameters(ctx, (*pb.BlockHash)(&bestblock))
			err_birk = json.Unmarshal([]byte(response_birk.Value), &bakerStatus)
			if err_birk != nil {
				fmt.Printf("Get BirkParameters:%#v \n", err_birk)
				return data, err_birk
			}
			// Search your baker number
			for i := 0; i < len(bakerStatus.Bakers); i++ {
				if bakerStatus.Bakers[i].BakerID == data.BakerId {
					data.BakerLotteryPower = bakerStatus.Bakers[i].BakerLotteryPower
					data.EstimatedBakingBlock = data.BakerLotteryPower * 35000
				}
			}
		}
	}
	return data, err
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	c := newconcordiumCollector()
	registry := prometheus.NewRegistry()
	registry.MustRegister(c)
	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

func main() {
	// Show help information
	flag.Usage = func() {
		const (
			usage = "Usage: concordium_exporter\n\n" + "Prometheus exporter for Concordium node metrics\n\n"
		)
		fmt.Fprint(flag.CommandLine.Output(), usage)
		flag.PrintDefaults()
		os.Exit(2)
	}

	var hport string

	// Fetch flag variables
	flag.StringVar(&url, "url", "localhost:10000", "Concordium gRPC URL")
	flag.StringVar(&hport, "hport", "9360", "The port listens on for HTTP requests")
	flag.StringVar(&password, "pwd", "rpcadmin", "The password to pass concordium node")
	flag.BoolVar(&bakerYn, "baker", false, "Whether your node is baking")

	// Fetch environment variables
	flag.VisitAll(func(f *flag.Flag) {
		if s := os.Getenv(strings.ToUpper("CCDEXPORTER_" + f.Name)); s != "" {
			f.Value.Set(s)
		}
	})
	flag.Parse()

	// Show each variables
	println("concordium-exporter: v" + version)
	println("Parameters:\nURL=" + url + "\nPORT=" + hport + "\nPASSWORD=" + password)
	println("Baker:", bakerYn)

	if len(flag.Args()) > 0 {
		flag.Usage()
	}

	// Start gRPC connection
	log.Printf("Initial connection started")
	conn, err := grpc.Dial(url, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("client connection error: %#v", err)
	}
	defer conn.Close()

	// Start collecting
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		metricsHandler(w, r)
	})
	log.Fatal(http.ListenAndServe(":"+hport, nil))
}
