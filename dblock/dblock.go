package dblock

/*
func (d *DomainBlock) Run() ResultInfo {
	pool := &Pool{
		WorkerCount: d.concurrency,
	}
	q := pool.Start()
	wg := sync.WaitGroup{}
	wg.Add(len(d.variations))

	go func(queue chan<- Task) {
		for i := 0; i < len(d.variations); i++ {
			domain := d.variations[i]

			queue <- func() {
				defer func() {
					recover()
					wg.Done()
				}()

				// this is a main domain
				if domain.How == "Main" && !d.checkMainDomain {
					return
				}

				ips, err := lookupNS(domain.Variant, d.dns)
				if err != nil {
					if !errors.Is(err, ErrorNoIPHost) {
						// DO something
						d.errorC <- domain.Variant
						return
					}
					d.outC <- DomainInfo{
						IsValid:    false,
						IsMain:     domain.Variant == d.domain,
						MainDomain: domain.MainDomain,
						Domain:     domain,
					}
					// NO A found
					return
				}
				hosts, err := lookupHost(domain.Variant, d.dns)
				if err != nil {
					if !errors.Is(err, ErrorNoIPHost) {
						// DO something
						d.errorC <- domain.Variant
						return
					}
					d.outC <- DomainInfo{
						IsValid:    false,
						IsMain:     domain.Variant == d.domain,
						MainDomain: domain.MainDomain,
						Domain:     domain,
					}
					// NO HOST found
					return
				}
				d.outC <- DomainInfo{
					IsValid:    true,
					IsMain:     domain.Variant == d.domain,
					Domain:     domain,
					MainDomain: domain.MainDomain,
					IP:         ips,
					Host:       hosts,
				}
			}
		}
		wg.Wait()
		pool.Stop()
		d.stop()

	}(q)

	return ResultInfo{
		Domain: d.domain,
		All:    len(d.variations),
		OutC:   d.outC,
		ErrC:   d.errorC,
	}
}

func New(cfg Config) *DomainBlock {
	return &DomainBlock{
		concurrency:     cfg.Concurrency,
		domain:          cfg.Domain,
		variations:      cfg.Variations,
		dns:             cfg.DNS,
		checkMainDomain: cfg.CheckMainDomain,
		Id:              uuid.NewString(),
		outC:            make(chan DomainInfo),
		errorC:          make(chan string),
	}
}
*/
