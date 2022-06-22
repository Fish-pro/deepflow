package example

var YamlDomainAliYun = []byte(`
# 名称
NAME: aliyun
# 云平台类型
TYPE: 9
CONFIG:
  # 所属区域标识
  region_uuid: ffffffff-ffff-ffff-ffff-ffffffffffff
  # 资源同步控制器
  controller_ip: 127.0.0.1
  # AccessKey ID
  # 阿里云控制台-accesskeys页面上获取用于API访问的密钥ID
  secret_id: xxxxxxxx
  # AccessKey Secret
  # 阿里云控制台-accesskeys页面上获取用于API访问的密钥KEY
  secret_key: xxxxxxx
  # 区域白名单，多个区域名称之间以英文逗号分隔
  include_regions:
  # 区域黑名单，多个区域名称之间以英文逗号分隔
  exclude_regions:
`)
