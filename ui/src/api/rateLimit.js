import request from '@/utils/request'

export function Add(data) {
  return request({
    url: '/api/rate-limit',
    method: 'post',
    data: data
  })
}

export function GetList(params) {
  return request({
    url: '/api/rate-limit',
    method: 'get',
    params: params
  })
}

export function Update(id, data) {
  return request({
    url: `/api/rate-limit/${id}`,
    method: 'put',
    data: data
  })
}

export function UpdateStatus(id) {
  return request({
    url: `/api/rate-limit/${id}/status`,
    method: 'put'
  })
}

export function Delete(id) {
  return request({
    url: `/api/rate-limit/${id}`,
    method: 'delete'
  })
}
