import request from '@/utils/request'

export function ListAttackLog(params) {
  return request({
    url: '/api/sampled-logs',
    method: 'get',
    params: params
  })
}
