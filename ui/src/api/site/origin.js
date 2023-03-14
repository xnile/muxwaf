import request from '@/utils/request'
import { method } from 'lodash'

export function GetOrigins(id) {
  return request({
    url: `/api/sites/${id}/origins`,
    method: 'get'
  })
}

export function UpdateOrigin(id, data) {
  return request({
    url: `/api/sites/origins/${id}`,
    method: 'put',
    data: data
  })
}

export function AddOrigins(id, data) {
  return request({
    url: `/api/sites/${id}/origins`,
    method: 'post',
    data: data
  })
}

export function DelOrigin(id) {
  return request({
    url: `/api/sites/origins/${id}`,
    method: 'delete'
  })
}
