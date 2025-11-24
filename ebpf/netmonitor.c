//go:build ignore
#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>
#include <linux/if_ether.h> // defines struct ethhdr, ETH_P_IP
#include <linux/ip.h>       // defines struct iphdr
#include <linux/in.h>       // optional: for IPPROTO_* constants
#include <linux/pkt_cls.h>   // for TC_ACT_OK
#include <bpf/bpf_endian.h>   // <- contains __bpf_ntohl, __bpf_htonl, etc.

struct {
    __uint(type, BPF_MAP_TYPE_ARRAY); 
    __type(key, __u32);
    __type(value, __u64);
    __uint(max_entries, 1);
} pkt_count SEC(".maps"); 

// count_packets atomically increases a packet counter on every invocation.
SEC("xdp") 
int count_packets() {
    __u32 key    = 0; 
    __u64 *count = bpf_map_lookup_elem(&pkt_count, &key); 
    if (count) { 
        __sync_fetch_and_add(count, 1); 
    }

    return XDP_PASS; 
}

struct {
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __type(key, __u32);
    __type(value, __u32);
    __uint(max_entries, 1);
} outgoing_addr SEC(".maps");

SEC("tc")
int monitor_egress(struct __sk_buff *skb)
{
    void *data     = (void *)(long)skb->data;
    void *data_end = (void *)(long)skb->data_end;

    struct ethhdr *eth = data;
    if ((void *)(eth + 1) > data_end)
        return TC_ACT_OK;

    if (eth->h_proto != __constant_htons(ETH_P_IP))
        return TC_ACT_OK;

    struct iphdr *iph = (void *)(eth + 1);
    if ((void *)(iph + 1) > data_end)
        return TC_ACT_OK;

    __u32 src = iph->saddr;
    __u32 dst = bpf_ntohl(iph->daddr);

    // increment counters for outgoing src IP
    if (dst != 0) {
      __u32 key = 0;
      bpf_map_update_elem(&outgoing_addr, &key, &dst, BPF_ANY);
    }

    return TC_ACT_OK;
}

char __license[] SEC("license") = "Dual MIT/GPL";
