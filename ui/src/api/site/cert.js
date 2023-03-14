import request from '@/utils/request'

export function GetCandidateCertificates(siteID, params) {
  return request({
    url: `/api/sites/${siteID}/certificates`,
    method: 'get',
    params: params
  })
}
