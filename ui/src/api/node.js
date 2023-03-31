import request from '@/utils/request'

export function ListNodes(params) {
  return request({
    url: '/api/nodes',
    method: 'get',
    params: params
  })
}

export function AddNode(data) {
  return request({
    url: '/api/nodes',
    method: 'post',
    data: data
  })
}

export function DelNode(id) {
  return request({
    url: `/api/nodes/${id}`,
    method: 'delete'
  })
}

export function SyncCfg(id) {
  return request({
    url: `/api/nodes/${id}/sync`,
    method: 'put'
  })
}

export function SwitchLogUpload(id) {
  return request({
    url: `/api/nodes/${id}/sample_log_upload`,
    method: 'put'
  })
}

export function SwitchStatus(id) {
  return request({
    url: `/api/nodes/${id}/status`,
    method: 'put'
  })
}
