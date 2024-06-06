etcdctl put /coreos.com/network/config '{"Network": "10.244.0.0/16", "SubnetLen": 24, "SubnetMin": "10.244.1.0","SubnetMax": "10.244.15.0", "Backend": {"Type": "vxlan"}}'
etcdctl get /coreos.com/network/config
flanneld -etcd-endpoints=http://127.0.0.1:2379 -etcd-prefix=/coreos.com/network -iface=eth0
