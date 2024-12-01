package main

import (
	"context"
	"flag"
	"github.com/ipfs/boxo/blockstore"
	exchange "github.com/ipfs/boxo/exchange"
	"github.com/ipfs/boxo/exchange/offline"
	"github.com/ipfs/boxo/filestore"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	leveldb "github.com/ipfs/go-ds-leveldb"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/ipfs/boxo/blockservice"
	"github.com/ipfs/boxo/examples/gateway/common"
	"github.com/ipfs/boxo/gateway"
)

var ex exchange.Interface

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	port := flag.Int("p", 8040, "port to run this gateway from")
	uploadPort := flag.Int("uploadP", 8041, "port to run this gateway from")
	flag.Parse()

	// Setups up tracing. This is optional and only required if the implementer
	// wants to be able to enable tracing.
	tp, err := common.SetupTracing(ctx, "File manager Gateway Example")
	if err != nil {
		log.Fatal(err)
	}
	defer (func() { _ = tp.Shutdown(ctx) })()

	blockService, err := newBlockService("fdb")
	if err != nil {
		log.Fatal(err)
	}

	// Creates the gateway API with the block service.
	backend, err := gateway.NewBlocksBackend(blockService)
	if err != nil {
		log.Fatal(err)
	}

	upload := func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("Access-Control-Allow-Origin", "*")

		body, err := io.ReadAll(req.Body)
		if err != nil {
			resp.WriteHeader(http.StatusInternalServerError)
			resp.Write([]byte(err.Error()))
		}
		block := blocks.NewBlock(body)
		cidv1 := cid.NewCidV1(cid.Raw, block.Cid().Hash())
		block, _ = blocks.NewBlockWithCid(body, cidv1)
		blockService.AddBlock(context.Background(), block)
		resp.WriteHeader(http.StatusOK)

		resp.Write([]byte(cidv1.String()))
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/upload", upload)
	//handler := common.NewUploadHandler(backend, upload)
	handler := common.NewHandler(backend)

	log.Printf("Listening on http://localhost:%d", *port)
	log.Printf("Metrics available at http://127.0.0.1:%d/debug/metrics/prometheus", *port)

	go func() {
		if err := http.ListenAndServe(":"+strconv.Itoa(*uploadPort), mux); err != nil {
			log.Fatal(err)
		}
	}()

	if err := http.ListenAndServe(":"+strconv.Itoa(*port), handler); err != nil {
		log.Fatal(err)
	}
}

func newBlockService(root string) (blockservice.BlockService, error) {
	lds, _ := leveldb.NewDatastore(root, nil)

	fm := filestore.NewFileManager(lds, root)
	fm.AllowFiles = true

	bs := blockstore.NewBlockstore(lds)
	fstore := filestore.NewFilestore(bs, fm)

	ex = offline.Exchange(bs)
	blockService := blockservice.New(fstore, ex)
	return blockService, nil
}
