import request from '@/utils/request'

// 分页获取IP白名单
export function GetIPList(params) {
  return request({
    url: '/api/whitelist/ip',
    method: 'get',
    params: params
  })
}

// 添加IP白名单
export function AddIP(data) {
  return request({
    url: '/api/whitelist/ip',
    method: 'post',
    data: data
  })
}

// 编辑IP白名单
export function UpdateIP(id, data) {
  return request({
    url: `/api/whitelist/ip/${id}`,
    method: 'put',
    data: data
  })
}

// 删除IP白名单
export function DeleteIP(id) {
  return request({
    url: `/api/whitelist/ip/${id}`,
    method: 'delete'
  })
}

// 更新IP白名单状态
export function UpdateIPStatus(id) {
  return request({
    url: `/api/whitelist/ip/${id}/status`,
    method: 'put'
  })
}

// 检测IP是否已经包含在白名单库中
export function IsIncluded(ip) {
  return request({
    url: '/api/whitelist/ip/isIncluded',
    method: 'get',
    params: {
      ip: ip
    }
  })
}
