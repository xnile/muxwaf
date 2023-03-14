import request from '@/utils/request'

// 分页获取URL白名单
export function GetURLList(params) {
  return request({
    url: '/api/whitelist/url',
    method: 'get',
    params: params
  })
}

// 添加URL白名单
export function AddURL(data) {
  return request({
    url: '/api/whitelist/url',
    method: 'post',
    data: data
  })
}

// 编辑URL白名单
export function UpdateURL(id, data) {
  return request({
    url: `/api/whitelist/url/${id}`,
    method: 'put',
    data: data
  })
}

// 删除URL白名单
export function DeleteURL(id) {
  return request({
    url: `/api/whitelist/url/${id}`,
    method: 'delete'
  })
}

// 更新URL白名单状态
export function UpdateURLStatus(id) {
  return request({
    url: `/api/whitelist/url/${id}/status`,
    method: 'put'
  })
}
