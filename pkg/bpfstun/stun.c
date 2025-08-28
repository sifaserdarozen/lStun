//go:build ignore
#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>
#include <linux/pkt_cls.h>      // struct __sk_buff
#include <linux/if_ether.h>     // struct ethhdr
#include <linux/ip.h>           // struct iphdr
#include <bpf/bpf_endian.h>     // bpf_ntohs

//#include <bpf/bpf_tracing.h>
//#include <linux/if_packet.h> 
//#include <linux/udp.h>
//#include <linux/bpf_common.h>

SEC("tc")
int ingress(struct __sk_buff *skb)
{
    bpf_printk("Incoming packet");
    void *data = (void *)(long)skb->data;
    void *data_end = (void *)(long)skb->data_end;

    struct ethhdr *eth = data;

    // Check eth data length
    if (data + sizeof(*eth) > data_end)
    {
        // return TC_ACT_SHOT;
        return TC_ACT_OK;
    }

    // Pass non ipv4 packets
    if (bpf_ntohs(eth->h_proto) != ETH_P_IP)
    {
        return TC_ACT_OK;
    }

    struct iphdr *ip4h = data + sizeof(*eth);

    // Checking ipv4 data length
    if (data + sizeof(*eth) + sizeof(*ip4h) > data_end)
    {
        // return TC_ACT_SHOT;
        return TC_ACT_OK;
    }

    // TODO: log some ip values here

    bpf_printk("%pI4 <---> %pI4", &ip4h->saddr, &ip4h->daddr);

    
    return TC_ACT_OK;
}

#if TEST

#endif // TEST


char __license[] SEC("license") = "GPL";