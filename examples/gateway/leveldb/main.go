package main

import (
	"context"
	"flag"
	"github.com/ipfs/boxo/blockstore"
	exchange "github.com/ipfs/boxo/exchange"
	"github.com/ipfs/boxo/exchange/offline"
	"github.com/ipfs/boxo/filestore"
	leveldb "github.com/ipfs/go-ds-leveldb"
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
	//uploadPort := flag.Int("uploadP", 8041, "port to run this gateway from")
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

	//mux := http.NewServeMux()
	//handler := common.NewUploadHandler(backend, upload)
	handler := common.NewHandler(backend, blockService)

	log.Printf("Listening on http://localhost:%d", *port)
	log.Printf("Metrics available at http://127.0.0.1:%d/debug/metrics/prometheus", *port)

	//go func() {
	//	if err := http.ListenAndServe(":"+strconv.Itoa(*uploadPort), mux); err != nil {
	//		log.Fatal(err)
	//	}
	//}()

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
