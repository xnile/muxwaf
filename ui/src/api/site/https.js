import request from '@/utils/request'

export function GetHttpsConfigs(siteID) {
  return request({
    url: `/api/sites/${siteID}/configs/https`,
    method: 'get'
  })
}
