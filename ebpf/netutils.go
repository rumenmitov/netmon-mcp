package ebpf

import (
    "log"
    "net"
    "os"
    "os/signal"
    "time"

    "github.com/cilium/ebpf/link"
)

func IncomingPacketsPerSecond() float32 {
    // load the compiled eBPF ELF into the kernel.
    var objs netmonObjects 
    if err := loadNetmonObjects(&objs, nil); err != nil {
        log.Fatal("Loading eBPF objects:", err)
    }
    defer objs.Close() 

    ifname := "wlp1s0" 
    iface, err := net.InterfaceByName(ifname)
    if err != nil {
        log.Fatalf("Getting interface %s: %s", ifname, err)
    }

    // attach count_packets to the network interface.
    xlink, err := link.AttachXDP(link.XDPOptions{ 
        Program:   objs.CountPackets,
        Interface: iface.Index,
    })
    if err != nil {
        log.Fatal("Attaching XDP:", err)
    }
    defer xlink.Close() 

    log.Printf("Counting incoming packets on %s..", ifname)

		// print out counter periodically
    tick := time.Tick(time.Second)
    stop := make(chan os.Signal, 5)
    signal.Notify(stop, os.Interrupt)

		i := 0
		var packets uint64 = 0

		loop:
    for {
        select {
        case <-tick:
            var count uint64
            err := objs.PktCount.Lookup(uint32(0), &count) 
            if err != nil {
                log.Fatal("Map lookup:", err)
            }

						packets += count

						i += 1

						if i >= 5 {
							break loop;
						}

        case <-stop:
            log.Print("Received signal, exiting..")
            return 0.0
        }
    }

		return float32(packets) / 5
}
