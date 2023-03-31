import request from '@/utils/request'

export function ListAttackLog(params) {
  return request({
    url: '/api/sample-logs',
    method: 'get',
    params: params
  })
}
