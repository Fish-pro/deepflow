package example

var YamlDomainKubernetes = []byte(`
# 名称
NAME: kubernetes
# 云平台类型
TYPE: 11
CONFIG:
  # 所属区域标识
  region_uuid: ffffffff-ffff-ffff-ffff-ffffffffffff
  # 资源同步控制器
  controller_ip: 127.0.0.1
  # POD子网IPv4地址最大掩码
  pod_net_ipv4_cidr_max_mask: 16
  # POD子网IPv6地址最大掩码
  pod_net_ipv6_cidr_max_mask: 64
  # 额外对接路由接口
  port_name_regex: ^(cni|flannel|cali|vxlan.calico|tunl)
`)
