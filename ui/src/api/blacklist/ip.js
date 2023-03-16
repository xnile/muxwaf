import request from '@/utils/request'

// 分页获取IP黑名单
export function ListBlacklistIP(params) {
  return request({
    url: '/api/blacklist/ip',
    method: 'get',
    params: params
  })
}

// 添加IP黑名单
export function InsertBlacklistIP(data) {
  return request({
    url: '/api/blacklist/ip',
    method: 'post',
    data: data
  })
}

// 编辑IP黑名单
export function UpdateBlacklistIP(id, data) {
  return request({
    url: `/api/blacklist/ip/${id}`,
    method: 'put',
    data: data
  })
}

// 删除IP黑名单
export function DeleteBlacklistIP(id) {
  return request({
    url: `/api/blacklist/ip/${id}`,
    method: 'delete'
  })
}

// 更新IP黑名单状态
export function UpdateBlacklistIPStatus(id) {
  return request({
    url: `/api/blacklist/ip/${id}/status`,
    method: 'put'
  })
}

export function IsIncluded(ip) {
  return request({
    url: '/api/blacklist/ip/isIncluded',
    method: 'get',
    params: {
      ip: ip
    }
  })
}
