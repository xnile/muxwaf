import request from '@/utils/request'

export function UpdateSiteRegionBlacklist(id, data) {
  return request({
    url: `/api/sites/${id}/region-blacklist`,
    method: 'put',
    data: data
  })
}

export function GetSiteRegionBlacklist(id) {
  return request({
    url: `/api/sites/${id}/region-blacklist`,
    method: 'get'
  })
}
