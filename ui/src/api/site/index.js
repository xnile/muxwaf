import request from '@/utils/request'

// const api = {
//   site: '/api/sites'
// }

export function AddSite(data) {
  return request({
    url: '/api/sites',
    method: 'post',
    data: data
  })
}

export function ListSite(params) {
  return request({
    url: '/api/sites',
    method: 'get',
    params: params
  })
}

export function GetALLSite() {
  return request({
    url: '/api/sites/all',
    method: 'get'
  })
}

export function UpdateSiteConfigs(id, data) {
  return request({
    url: `/api/sites/${id}/configs`,
    method: 'put',
    data: data
  })
}

export function UpdateSiteHttps(id, data) {
  return request({
    url: `/api/sites/${id}/https`,
    method: 'put',
    data: data
  })
}

export function GetConfigs(id) {
  return request({
    url: `/api/sites/${id}/configs`,
    method: 'get'
  })
}

export function DelSite(id) {
  return request({
    url: `/api/sites/${id}`,
    method: 'delete'
  })
}

export function UpdateStatus(id) {
  return request({
    url: `/api/sites/${id}/status`,
    method: 'put'
  })
}
