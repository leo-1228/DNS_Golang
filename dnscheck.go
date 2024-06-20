package dnscheck

import (
	"dnscheck/api"
	"dnscheck/dblock"
	"dnscheck/dbs"
	"dnscheck/logger"
	"dnscheck/reader"
	"dnscheck/report"
	"dnscheck/saver"
	"dnscheck/utils"
	"dnscheck/variation"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/exp/slices"
)

type DomainBatcher struct {
	BatchChan     chan []*dbs.DomainBlock
	rdr           *reader.DbReader
	varController *variation.Variation
	cfg           *Config
}

func NewDomainBatcher(rdr *reader.DbReader, varController *variation.Variation, cfg *Config) *DomainBatcher {
	batchSize := cfg.PreloadSize
	if batchSize == 0 {
		batchSize = 5
	}
	return &DomainBatcher{
		BatchChan:     make(chan []*dbs.DomainBlock, batchSize),
		rdr:           rdr,
		varController: varController,
		cfg:           cfg,
	}
}

// Moved to new duplicate check
func (db *DomainBatcher) CheckForDuplicatesInDB(domains []*dbs.DomainBlock) (int, []*dbs.DomainBlock) {

	numDuplicates := 0
	for _, dblock := range domains {
		variations := make([]variation.VariationRecord, 0)
		whereClause := make([]string, 10000)
		allConflicts := make([]string, 0)
		for i, v := range dblock.Variations {

			whereClause[i%10000] = v.Variant

			if i > 0 && i%10000 == 0 {

				conflicts, err := dbs.Service.DomainDuplicateCheck(whereClause)
				if err != nil {
					// logger.Error("Duplicate check returned an error: ", err)
				}
				allConflicts = append(allConflicts, conflicts...)
			}

			if i == len(dblock.Variations)-1 {
				conflicts, err := dbs.Service.DomainDuplicateCheck(whereClause[:i%10000])
				if err != nil {
					// logger.Error("Duplicate check returned an error: ", err)
				}
				allConflicts = append(allConflicts, conflicts...)
			}

		}
		for _, v := range dblock.Variations {
			if !slices.Contains(allConflicts, v.Variant) {
				variations = append(variations, v)
			} else {
				numDuplicates++
			}
		}

		dblock.Variations = variations
	}

	return numDuplicates, domains
	// return dbs.VariationDomainService.AddVariationDomains(domains)
}

// Moved to new duplicate check
func (db *DomainBatcher) CheckForDuplicates(domains []*dbs.DomainBlock) (int, []*dbs.DomainBlock) {
	existingDomains, _ := dbs.Service.GetDomainsForDuplicateCheck()
	numDuplicates := 0
	for _, dblock := range domains {
		variations := make([]variation.VariationRecord, 0)
		for _, v := range dblock.Variations {
			if !slices.Contains(existingDomains, v.Variant) {
				variations = append(variations, v)
			} else {
				numDuplicates++
			}
		}
		dblock.Variations = variations
	}

	return numDuplicates, domains
	// return dbs.VariationDomainService.AddVariationDomains(domains)
}

func (db *DomainBatcher) Run() error {
	// firstRun := true
	for {

		baseDomains, err := db.rdr.Batch()
		if err != nil {
			if errors.Is(err, reader.ErrRangeReached) {
				logger.Info("success: all domains in the given range have been processed.")
				return nil
			}
			if errors.Is(err, reader.ErrDomainsEOF) {
				logger.Info("success: all domains in the given source file have been processed.")
				return nil
			}
			logger.Error("error: ", err.Error())
			return nil
		}

		logger.Info("calculating variations...")
		allVariations, err := db.varController.Run(baseDomains)
		if err != nil {
			logger.Error("error: ", err.Error())
			return nil
		}
		c := 0
		for k := range allVariations {
			c += len(allVariations[k].Variations)
		}

		logger.Info("variations calculated for domains: ", strings.Join(baseDomains, ","), ". Got ", c, " entries. Checking for duplicates...")

		// TODO: Check for duplicates?

		domainBlocks := make([]*dbs.DomainBlock, 0)
		// Wait depends on the number of clients
		// Do this iteration via channels and clients
		for i := 0; i < len(allVariations); i++ {

			db := &dbs.DomainBlock{
				Concurrency:     db.cfg.Concurrency,
				Domain:          allVariations[i].Domain,
				Variations:      allVariations[i].Variations,
				Dns:             db.cfg.DNSes,
				CheckMainDomain: db.cfg.CheckMainDomains,
				Id:              uuid.NewString(),
			}

			domainBlocks = append(domainBlocks, db)
		}
		if db.cfg.CheckDuplication {

			if db.cfg.DuplicateDetectionMode == "in-memory" {
				duplicates, domainBlocks := db.CheckForDuplicates(domainBlocks)
				logger.Info("Adding domainblock with batch_size ", len(domainBlocks), " domains to queue. Found ", duplicates, " duplicates through in-memory scan")
			} else {
				duplicates, domainBlocks := db.CheckForDuplicatesInDB(domainBlocks)
				logger.Info("Adding domainblock with batch_size ", len(domainBlocks), " domains to queue. Found ", duplicates, " duplicates through sqlite scan")
			}

		} else {
			logger.Info("Adding domainblock with batch_size ", len(domainBlocks), " domains to queue.")
		}

		// domainBlocksToDbBlocks
		blocks := make([]dbs.DbBlock, len(domainBlocks))
		// TODO: DomainBlockToDbBlock
		for i, v := range domainBlocks {
			dbBlock := dbs.DbBlock{
				Concurrency:     v.Concurrency,
				CheckMainDomain: v.CheckMainDomain,
				Domain:          v.Domain,
				Dns:             strings.Join(v.Dns, ";"),
				Variants:        make([]dbs.DbVariants, len(v.Variations)),
			}

			for j, v2 := range v.Variations {
				dbBlock.Variants[j] = dbs.DbVariants{
					Domain:     v2.Variant,
					MainDomain: v2.MainDomain,
				}
			}
			blocks[i] = dbBlock
		}
		err = dbs.VariantService.InsertVariants(blocks)
		if err != nil {
			logger.Error("Failed to insert variants: ", err)
		}

		// TODO: Fix this first run issue
		/*if firstRun {

			select {
			case db.BatchChan <- domainBlocks:
				continue
			default:
				logger.Info("Resetting unfinished domains")
				err := dbs.Service.ResetUnfinishedDomains()
				if err != nil {
					logger.Error("Failed to reset unfinished domains: ", err)
				}
				firstRun = false
			}

		} else {
			db.BatchChan <- domainBlocks
		}*/

	}
}

var ops uint64
var respTime uint64
var valids uint64

func handleBlockPart(d *dbs.DomainBlock, start, end int, dns string, trackTimes bool) ([]dbs.DomainInfo, []string) {
	// logrus.Info("Handling block from ", start, " to ", end)
	infos := make([]dbs.DomainInfo, 0)
	errs := make([]string, 0)

	for i := start; i < end; i++ {

		//if i > 0 && i%100 == 0 {
		//	logrus.Warn("Block ", start, " finished 100 domain checks")
		//}

		domain := d.Variations[i]
		// this is a main domain
		if domain.How == "Main" && !d.CheckMainDomain {
			continue
		}

		if domain.MainDomain == "" {
			continue
		}

		atomic.AddUint64(&ops, 1)

		ips, err := dblock.LookupNS(domain.Variant, dns, trackTimes)
		if err != nil {
			if !errors.Is(err, dblock.ErrorNoIPHost) {
				// logrus.Info("Nameserver error: ", err)
				// TODO: This might be configurable later
				//errs = append(errs, domain.Variant)
				continue
			}
			/*infos = append(infos, dbs.DomainInfo{
				IsValid:    false,
				IsMain:     domain.Variant == d.Domain,
				MainDomain: domain.MainDomain,
				Domain:     domain,
			})*/
			// logrus.Info("Invalid nameserver: ", err)
			// NO A found
			continue
		}
		hosts, err := dblock.LookupHost(domain.Variant, dns)
		if err != nil {
			if !errors.Is(err, dblock.ErrorNoIPHost) {
				// DO something
				// TODO: This might be configurable later
				// errs = append(errs, domain.Variant)
				continue
			}
			/*infos = append(infos, dbs.DomainInfo{
				IsValid:    false,
				IsMain:     domain.Variant == d.Domain,
				MainDomain: domain.MainDomain,
				Domain:     domain,
			})*/
			// logger.Error("Invalid host for ", domain, ": ", err)
			// NO HOST found
			continue
		}
		infos = append(infos, dbs.DomainInfo{
			IsValid:    true,
			IsMain:     domain.Variant == d.Domain,
			Domain:     domain,
			MainDomain: domain.MainDomain,
			IP:         ips,
			Host:       hosts,
		})
		atomic.AddUint64(&valids, 1)
		// logger.Info("Valid domain found: ", domain.Variant)
	}
	return infos, errs
}

func StartClient(serverUrl, secret, clientId string) {

	clientApi := api.NewApiV1Client(serverUrl, secret, clientId)
	sleepDur := 500 * time.Millisecond

	err := dbs.SetupClientCache()
	if err != nil {
		log.Fatal(err)
	}

	for {
		logger.Info("Fetching config from server...")
		conf, err := clientApi.GetConfig()
		logger.Info("Done fetching config from server")
		if err != nil {
			logger.Error("Failed to obtain config from server: ", err)
			time.Sleep(sleepDur)
			continue
		}

		cachedResults, err := dbs.ClientCache.GetCachedResults()
		if err != nil {
			logger.Error("Failed to fetch cached results: ", err)
		}
		for _, r := range cachedResults {
			logger.Info("Sending cached results to server")
			r.ClientId = conf.ClientId
			err = clientApi.SetResults(r)
			if err != nil {
				logger.Error("Failed to send cached result to server: ", err)
				continue
			}

			dbs.ClientCache.RemoveResult(&r)
		}

		if len(conf.DomainBlocks) == 0 {
			logger.Info("No DomainBlock received for processing")
			time.Sleep(sleepDur)
			continue
		}

		for _, d := range conf.DomainBlocks {

			logger.Info("Got block with domain ", d.Domain)

			// TODO: This could be more in the future
			resultInfo := dbs.ResultInfo{
				DomainBlockId: d.Id,
				Domain:        d.Domain,
			}

			var wg sync.WaitGroup
			m := sync.Mutex{}

			// Specific configuration overwrites default one
			concurrency := d.Concurrency
			if conf.Concurrency > 0 {
				concurrency = conf.Concurrency
			}

			dnses := d.Dns
			if conf.DNS != nil && len(conf.DNS) > 0 {
				dnses = conf.DNS
			}

			partLen := utils.CeilForce(len(d.Variations), concurrency)
			dnsIndex := 0
			logger.Info("Checking ", len(d.Variations), " domains with concurrency ", concurrency)

			if len(d.Variations) != 0 {
				ops = 0
				valids = 0
				respTime = 0

				stopChan := make(chan bool)
				go func() {
					t := time.Now()
					for {
						select {
						case <-stopChan:
							return
						default:
							secs := time.Since(t)
							//ClearMultiLine := "\033[" + fmt.Sprint(1) + "A"
							//fmt.Print(ClearMultiLine)

							// fmt.Printf("\rOn %d/10", 1)

							logger.Info("Processed ", ops, " / ", len(d.Variations), " domains in ", fmt.Sprintf("%.2f", secs.Seconds()), " seconds (", valids, ") valid")
							time.Sleep(5 * time.Second)
						}
					}
				}()

				for i := 0; i < concurrency; i++ {
					start := partLen * i
					end := utils.Min(start+partLen, len(d.Variations))
					dns := dnses[dnsIndex%len(dnses)]
					wg.Add(1)
					go func(start, end int, dns string) {
						res, errs := handleBlockPart(d, start, end, dns, conf.TrackResponseTimes)
						m.Lock()
						resultInfo.Infos = append(resultInfo.Infos, res...)
						resultInfo.Errors = append(resultInfo.Errors, errs...)
						m.Unlock()
						wg.Done()

					}(start, end, dns)
					dnsIndex++
				}
				wg.Wait()
				stopChan <- true
			}

			logger.Info("Processing of block done. Got ", valids, "/", len(d.Variations), " valid domains")
			resultInfo.ClientId = conf.ClientId
			logger.Info("Sending back results for domain ", resultInfo.Domain, "...")
			err = clientApi.SetResults(resultInfo)
			logger.Info("Sent back results for domain ", resultInfo.Domain)
			if err != nil {
				// logger.Error("Failed to send back results, caching now: ", err)
				dbs.ClientCache.AddResult(&resultInfo)
				logger.Info("Cached results")

			}
		}

		time.Sleep(sleepDur)
	}
}

func randomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length+2)
	rand.Read(b)
	return fmt.Sprintf("%x", b)[2 : length+2]
}

func StartServer(cfg *Config, jwtSecret string) {

	runId := randomString(32)
	logger.Info("Starting with runId=", runId)
	rangFrom, rangeTo, _ := cfg.Records()
	// Opt-in to get Domanis out of the database instead of the config file
	// rd, err := reader.New(reader.Config{
	rd, err := reader.NewDbReader(reader.Config{
		From:            rangFrom,
		To:              rangeTo,
		Workspace:       cfg.Workspace,
		BatchSize:       cfg.BatchSize,
		DomainsFileName: cfg.DomainsFile,
	})
	if err != nil {
		logger.Error("error: ", err.Error())
		return
	}

	err = dbs.InitializeDatabaseService(cfg.MainDBType, cfg.MainDBSettings)
	if err != nil {
		logger.Error("error: ", err.Error())
		return
	}

	err = dbs.InitializeInvalidDomainDatabaseService()
	if err != nil {
		logger.Error("error: ", err.Error())
		return
	}

	dbs.ClientDB.SpawnClientConfig = func(id string) *dbs.ClientConfig {

		clientId := id
		if id == "new" {
			clientId = uuid.NewString()
		}

		conf := &dbs.ClientConfig{
			ClientId:         clientId,
			PureLogMode:      cfg.PureLogMode,
			DNS:              cfg.DNSes,
			CheckMainDomains: cfg.CheckMainDomains,
			Concurrency:      cfg.Concurrency,
		}

		if cfg.Clients != nil && id != "new" {
			clientConf, ok := cfg.Clients[id]
			if ok {
				// merge configs
				conf.CheckMainDomains = clientConf.CheckMainDomains
				conf.Concurrency = clientConf.Concurrency
				conf.DNS = clientConf.DNS
				conf.TrackResponseTimes = clientConf.TrackResponseTimes
			}
		}

		return conf
	}

	// READ dict files
	dict, err := cfg.LoadDictionary()
	if err != nil {
		logger.Error("error: ", err.Error())
		return
	}

	// TODO: Rethink if we need this again...
	/*mainDomains, err := cfg.LoadMainDomains()
	if err != nil {
		logger.Error("error: ", err.Error())
		return
	}
	if len(mainDomains) > 0 {
		tmp := make([]string, 0)
		for _, m := range mainDomains {
			info, err := variation.DomainInfo(m, cfg.Spaces)
			if err != nil {
				logger.Error("error: ", err.Error())
				return
			}
			tmp = append(tmp, info.GetInAllSpaces(cfg.Spaces)...)
		}
		mainDomains = append(mainDomains, tmp...)
	}*/

	logger.Info("Automatic update mechanism: Checking for entries to add to TblProcessedDomains. This may take a while.")
	res, err := dbs.Service.AddNewEntriesToTableProcessedDomains()
	if err != nil {
		logger.Warning("Automatic update mechanism: Failed to add new entries to TblProcessedDomains: ", err)
	} else {
		logger.Info("Automatic update mechanism: Added ", res, " new domains to TblProcessedDomains.")
	}

	// Set a default port if "PORT" is not set
	apiUrl := cfg.ApiUrl
	if cfg.ApiUrl == "" {
		apiUrl = ":8810"
	}

	// Create a new Gin router
	router := gin.New()
	router.Use(cors.Default())

	// Define a route handler for the root URL ("/")
	router.GET("/newdomains/all", func(c *gin.Context) {

		domains, err := dbs.Service.GetAllNewDomains()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Could not get all domains",
			})
			logger.Error("API: Could not get all domains: ", err)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"data": domains,
		})
	})

	router.GET("/newdnsdata", func(c *gin.Context) {

		domains, err := dbs.Service.GetAllNewDomainsFromDNSData()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Could not get all domains",
			})
			logger.Error("API: Could not get new dnsdata: ", err)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"data": domains,
		})
	})

	// Define a route handler for the root URL ("/")
	router.GET("/newdomains/current", func(c *gin.Context) {
		domains, err := dbs.Service.GetCurentNewDomains(runId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Could not get current domains",
			})
			logger.Error("API: Could not get current domains: ", err)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"data": domains,
		})
	})

	go func() {
		err = router.Run(apiUrl)
		log.Fatal(err)
	}()

	variationController, err := variation.New(variation.Config{
		MainDomains:                make([]string, 0),
		CheckMainDomainDuplication: cfg.CheckDuplication,
		ValidSpaces:                cfg.Spaces,
		HomoglyphMethod:            cfg.Homoglyph(),
		HomoglyphNormal:            !cfg.HomoglyphDouble,
		AdditionMethod:             cfg.Addition(),
		BitsquattingMethod:         cfg.Bitsquatting(),
		HyphenationMethod:          cfg.Hyphenation(),
		InsertionMethod:            cfg.Insertion(),
		OmissionMethod:             cfg.Omission(),
		RepetitionMethod:           cfg.Repetition(),
		ReplacementMethod:          cfg.Replacement(),
		TranspositionMethod:        cfg.Transposition(),
		VowelSwapMethod:            cfg.VowelSwap(),
		DictionaryMethod:           cfg.Dictionary(),
		DictionaryData:             dict,
	})
	if err != nil {
		logger.Error("unknown error in variation ", err.Error())
		return
	}

	saveController, err := saver.New(cfg.Workspace, cfg.SaveInvalidDomains)
	if err != nil {
		logger.Error("saver: ", err.Error())
		return
	}

	reporter := report.New(cfg.LogMode())
	go func() {
		apiServer := api.NewApiV1Server(cfg.Url, jwtSecret)
		log.Fatal(apiServer.Serve())
	}()

	dBatcher := NewDomainBatcher(rd, variationController, cfg)

	go func() {
		err := dBatcher.Run()
		if err != nil {
			log.Fatal(err)
		}

	}()

	dbs.ClientDB.GetBatchForClient = func(id string) []*dbs.DomainBlock {
		// Get number of blocks for client id
		// Get entries from DB
		cl, ok := cfg.Clients[id]
		batchSize := cfg.BatchSize
		if ok {
			batchSize = cl.BatchSize
		}

		vrs, err := dbs.VariantService.GetVariants(batchSize)
		if err != nil {
			//TODO: WHat here
		}
		domainBlocks := make([]*dbs.DomainBlock, len(vrs))

		for i, v := range vrs {
			dBlock := &dbs.DomainBlock{
				Concurrency:     v.Concurrency,
				CheckMainDomain: v.CheckMainDomain,
				Domain:          v.Domain,
				Dns:             strings.Split(v.Dns, ";"),
				Variations:      make([]variation.VariationRecord, len(v.Variants)),
			}

			for j, v2 := range v.Variants {
				variant := variation.VariationRecord{
					Variant:    v2.Domain,
					MainDomain: v2.MainDomain,
				}

				dBlock.Variations[j] = variant
			}

			domainBlocks[i] = dBlock
		}

		// TODO: DbBlock to DomainBlock
		return domainBlocks
		// return <-dBatcher.BatchChan

	}

	logger.Info("Started dnscheck-server. Waiting for client requests...")

	go func() {
		// TODO: Reset started ProcessedDomains that are not done after 24h
	}()

	// numBlocks := 0
	for {

		//if numBlocks%cfg.BatchSize == 0 {
		//	reporter.NewScreen("new batch started", 10)
		//}

		dblocks := make([]dbs.ResultInfo, 0)
		// logger.Info("Waiting for completion")
		res := dbs.ClientDB.WaitForClientCompletion()
		logger.Info("Completed block with domain ", res.Domain)

		for {
			// Store back to sqlite to not check again
			err = dbs.Service.MarkDomainProcessed(&dbs.ProcessedDomain{
				Domain:     res.Domain,
				Variations: strings.Join(cfg.Actions, ","),
				//NewlyAdded: true,
				//RunId:      runId,
			})
			if err != nil {

				if strings.Contains(err.Error(), "SQLITE_BUSY") || strings.Contains(err.Error(), "cannot start a transaction within a transaction") {
					time.Sleep(3 * time.Second)
					continue
				} else {
					logger.Error(err)
				}
			}
			break
		}

		for {
			// Store back to sqlite to not check again
			err = dbs.VariantService.MarkMainDomainProcessed(res.Domain)
			if err != nil {

				if strings.Contains(err.Error(), "SQLITE_BUSY") || strings.Contains(err.Error(), "cannot start a transaction within a transaction") {
					time.Sleep(3 * time.Second)
					continue
				} else {
					logger.Error(err)
				}
			}
			break
		}

		// Save all
		//err = dbs.VariationDomainService.MarkVariationsProcessed(res.Domain)
		//if err != nil {
		// Ignore this, it is not important
		// logger.Error(err)
		//}

		for _, v := range res.Infos {

			nameServers := strings.Join(v.IP, ",")
			aRecords := strings.Join(v.Host, ",")

			if !v.IsValid {
				continue
			}
			/*if !v.IsValid {
				err = dbs.InvalidDomainService.AddInvalidDomain(&dbs.DbVariationDomain{
					Domain: v.Domain.Variant,
				})

				// TODO: What here? ...
				if err != nil {
					if strings.Contains(strings.ToLower(err.Error()), "constraint failed") {
						// logger.Normal("Domain ", v.Domain.Variant, " already found in database")
						continue
					} else {
						logger.Error(err)
					}
				}
				continue
			}*/
			for {
				err = dbs.Service.AddValidDomainBlock(&dbs.DNSDataResult{
					DomainName:  v.Domain.Variant,
					ARecords:    aRecords,
					Nameservers: nameServers,
				})

				// TODO: What here? ...
				if err != nil {
					if strings.Contains(strings.ToLower(err.Error()), "constraint failed") {

					} else if strings.Contains(err.Error(), "SQLITE_BUSY") || strings.Contains(err.Error(), "cannot start a transaction within a transaction") {
						time.Sleep(3 * time.Second)
						continue
					} else if strings.Contains(err.Error(), "Domain already exists") {

					} else {
						logger.Error(err)
					}
				}

				break
			}

			for {
				// Store back to sqlite to not check again
				err = dbs.Service.MarkDomainProcessed(&dbs.ProcessedDomain{
					Domain: v.Domain.Variant,
					// Variations: strings.Join(cfg.Actions, ","),
					NewlyAdded: true,
					RunId:      runId,
				})
				if err != nil {

					if strings.Contains(err.Error(), "SQLITE_BUSY") || strings.Contains(err.Error(), "cannot start a transaction within a transaction") {
						time.Sleep(3 * time.Second)
						continue
					} else {
						logger.Error(err)
					}
				}
				break
			}
		}

		/*for _, v := range res.Errors {

			err = dbs.VariationDomainService.AddVariationDomain(&dbs.DbVariationDomain{
				Domain: v,
			})

			// TODO: What here? ...
			if err != nil {
				if strings.Contains(strings.ToLower(err.Error()), "constraint failed") {
					// logger.Normal("Domain ", v.Domain.Variant, " already found in database")
					continue
				} else {
					logger.Error(err)
				}
			}
		}*/

		// blocking
		saveController.Run(dblocks, reporter)

		//if numBlocks > 0 && numBlocks%cfg.BatchSize == 0 {
		//		reporter.NewBatch()
		//	}
	}

	// Wait
	// dblocks = append(dblocks, resultInfo)

}
